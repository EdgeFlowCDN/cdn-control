package middleware

import (
	"bytes"
	"context"
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditLog is a Gin middleware that logs all non-GET requests to the audit_logs table.
func AuditLog(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only log mutating requests
		if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" || c.Request.Method == "HEAD" {
			c.Next()
			return
		}

		// Read body for audit details (then restore it)
		var details []byte
		if c.Request.Body != nil {
			details, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(details))
		}

		// Process the request
		c.Next()

		// Extract user info set by JWTAuth middleware
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		uid, _ := userID.(int64)
		uname, _ := username.(string)

		action := c.Request.Method
		resource := c.FullPath()
		if resource == "" {
			resource = c.Request.URL.Path
		}
		resourceID := c.Param("id")
		if resourceID == "" {
			resourceID = c.Param("cid")
		}
		if resourceID == "" {
			resourceID = c.Param("oid")
		}
		if resourceID == "" {
			resourceID = c.Param("rid")
		}

		ip := c.ClientIP()

		// Store as valid JSON; fall back to null if body is empty
		var detailsJSON interface{}
		if len(details) > 0 {
			detailsJSON = string(details)
		}

		_, err := db.Exec(context.Background(),
			`INSERT INTO audit_logs (user_id, username, action, resource, resource_id, details, ip)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uid, uname, action, resource, resourceID, detailsJSON, ip,
		)
		if err != nil {
			log.Printf("audit: failed to write log: %v", err)
		}
	}
}
