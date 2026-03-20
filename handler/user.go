package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EdgeFlowCDN/cdn-control/middleware"
	"github.com/EdgeFlowCDN/cdn-control/model"
)

type UserHandler struct {
	db *pgxpool.Pool
}

func NewUserHandler(db *pgxpool.Pool) *UserHandler {
	return &UserHandler{db: db}
}

type CreateUserReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role == "" {
		req.Role = "user"
	}

	hash, err := middleware.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	var user model.User
	err = h.db.QueryRow(context.Background(),
		`INSERT INTO users (username, password, role) VALUES ($1, $2, $3)
		 RETURNING id, username, role, created_at`,
		req.Username, hash, req.Role,
	).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) List(c *gin.Context) {
	rows, err := h.db.Query(context.Background(),
		"SELECT id, username, role, created_at FROM users ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, u)
	}
	if users == nil {
		users = []model.User{}
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var user model.User
	err := h.db.QueryRow(context.Background(),
		"SELECT id, username, role, created_at FROM users WHERE id = $1", userID,
	).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	var currentHash string
	err := h.db.QueryRow(context.Background(),
		"SELECT password FROM users WHERE id = $1", userID,
	).Scan(&currentHash)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if !middleware.CheckPassword(req.OldPassword, currentHash) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "old password is incorrect"})
		return
	}

	newHash, err := middleware.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	_, err = h.db.Exec(context.Background(),
		"UPDATE users SET password = $1 WHERE id = $2", newHash, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}
