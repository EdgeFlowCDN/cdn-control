package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditHandler struct {
	db *pgxpool.Pool
}

func NewAuditHandler(db *pgxpool.Pool) *AuditHandler {
	return &AuditHandler{db: db}
}

type AuditLog struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id"`
	Details    *string   `json:"details"`
	IP         string    `json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
}

func (h *AuditHandler) List(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	user := c.Query("user")

	query := "SELECT id, user_id, username, action, resource, resource_id, details, ip, created_at FROM audit_logs WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from time format, use RFC3339"})
			return
		}
		query += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, t)
		argIdx++
	}

	if to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to time format, use RFC3339"})
			return
		}
		query += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, t)
		argIdx++
	}

	if user != "" {
		query += fmt.Sprintf(" AND username = $%d", argIdx)
		args = append(args, user)
		argIdx++
	}

	query += " ORDER BY id DESC LIMIT 100"

	rows, err := h.db.Query(context.Background(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Username, &l.Action, &l.Resource, &l.ResourceID, &l.Details, &l.IP, &l.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		logs = append(logs, l)
	}
	if logs == nil {
		logs = []AuditLog{}
	}

	c.JSON(http.StatusOK, logs)
}
