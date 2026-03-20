package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EdgeFlowCDN/cdn-control/middleware"
	"github.com/EdgeFlowCDN/cdn-control/model"
)

type AuthHandler struct {
	db          *pgxpool.Pool
	expireHours int
}

func NewAuthHandler(db *pgxpool.Pool, expireHours int) *AuthHandler {
	return &AuthHandler{db: db, expireHours: expireHours}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	err := h.db.QueryRow(context.Background(),
		"SELECT id, username, password, role FROM users WHERE username = $1",
		req.Username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !middleware.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role, h.expireHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, model.LoginResp{Token: token})
}

// InitAdmin creates a default admin user if none exists.
func (h *AuthHandler) InitAdmin() error {
	var count int
	err := h.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM users WHERE role = 'admin'",
	).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	hash, err := middleware.HashPassword("admin123")
	if err != nil {
		return err
	}
	_, err = h.db.Exec(context.Background(),
		"INSERT INTO users (username, password, role) VALUES ($1, $2, $3)",
		"admin", hash, "admin",
	)
	return err
}
