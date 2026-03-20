package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookHandler struct {
	db *pgxpool.Pool
}

func NewWebhookHandler(db *pgxpool.Pool) *WebhookHandler {
	return &WebhookHandler{db: db}
}

type Webhook struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateWebhookReq struct {
	URL    string   `json:"url" binding:"required"`
	Events []string `json:"events" binding:"required"`
}

func (h *WebhookHandler) Create(c *gin.Context) {
	var req CreateWebhookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	var wh Webhook
	err := h.db.QueryRow(context.Background(),
		`INSERT INTO webhooks (user_id, url, events, active)
		 VALUES ($1, $2, $3, true)
		 RETURNING id, user_id, url, events, active, created_at`,
		userID, req.URL, req.Events,
	).Scan(&wh.ID, &wh.UserID, &wh.URL, &wh.Events, &wh.Active, &wh.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create webhook: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wh)
}

func (h *WebhookHandler) List(c *gin.Context) {
	userID, _ := c.Get("user_id")

	rows, err := h.db.Query(context.Background(),
		"SELECT id, user_id, url, events, active, created_at FROM webhooks WHERE user_id = $1 ORDER BY id",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var webhooks []Webhook
	for rows.Next() {
		var wh Webhook
		if err := rows.Scan(&wh.ID, &wh.UserID, &wh.URL, &wh.Events, &wh.Active, &wh.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		webhooks = append(webhooks, wh)
	}
	if webhooks == nil {
		webhooks = []Webhook{}
	}

	c.JSON(http.StatusOK, webhooks)
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID, _ := c.Get("user_id")

	tag, err := h.db.Exec(context.Background(),
		"DELETE FROM webhooks WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// SendWebhook sends a webhook event to all registered URLs matching the event.
// It retries up to 3 times with 1s backoff on failure.
func SendWebhook(db *pgxpool.Pool, event string, payload interface{}) {
	rows, err := db.Query(context.Background(),
		"SELECT id, url FROM webhooks WHERE active = true AND $1 = ANY(events)", event)
	if err != nil {
		log.Printf("webhook: failed to query webhooks: %v", err)
		return
	}
	defer rows.Close()

	type target struct {
		id  int64
		url string
	}
	var targets []target
	for rows.Next() {
		var t target
		if err := rows.Scan(&t.id, &t.url); err != nil {
			log.Printf("webhook: failed to scan row: %v", err)
			continue
		}
		targets = append(targets, t)
	}

	body, err := json.Marshal(gin.H{
		"event":     event,
		"payload":   payload,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		log.Printf("webhook: failed to marshal payload: %v", err)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, t := range targets {
		go func(url string) {
			var lastErr error
			for attempt := 0; attempt < 3; attempt++ {
				if attempt > 0 {
					time.Sleep(time.Duration(attempt) * time.Second)
				}
				resp, err := client.Post(url, "application/json", bytes.NewReader(body))
				if err != nil {
					lastErr = err
					continue
				}
				resp.Body.Close()
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					return
				}
				lastErr = nil
			}
			if lastErr != nil {
				log.Printf("webhook: failed to deliver to %s after 3 attempts: %v", url, lastErr)
			}
		}(t.url)
	}
}
