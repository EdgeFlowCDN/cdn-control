package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EdgeFlowCDN/cdn-control/model"
)

type PurgeHandler struct {
	db *pgxpool.Pool
}

func NewPurgeHandler(db *pgxpool.Pool) *PurgeHandler {
	return &PurgeHandler{db: db}
}

func (h *PurgeHandler) PurgeURL(c *gin.Context) {
	h.createPurgeTask(c, "url")
}

func (h *PurgeHandler) PurgeDir(c *gin.Context) {
	h.createPurgeTask(c, "dir")
}

func (h *PurgeHandler) PurgeAll(c *gin.Context) {
	var req model.PurgeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// For purge all, targets is just the domain
	req.Targets = []string{"*"}
	h.createTask(c, "all", req)
}

func (h *PurgeHandler) Prefetch(c *gin.Context) {
	var req model.PrefetchReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.createTask(c, "prefetch", model.PurgeReq{
		Targets: req.URLs,
		Domain:  req.Domain,
	})
}

func (h *PurgeHandler) GetTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	var task model.PurgeTask
	err = h.db.QueryRow(context.Background(),
		`SELECT id, type, targets, domain, status, created_at, completed_at
		 FROM purge_tasks WHERE id = $1`, id,
	).Scan(&task.ID, &task.Type, &task.Targets, &task.Domain, &task.Status, &task.CreatedAt, &task.CompletedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *PurgeHandler) createPurgeTask(c *gin.Context, purgeType string) {
	var req model.PurgeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.createTask(c, purgeType, req)
}

func (h *PurgeHandler) createTask(c *gin.Context, taskType string, req model.PurgeReq) {
	var task model.PurgeTask
	err := h.db.QueryRow(context.Background(),
		`INSERT INTO purge_tasks (type, targets, domain, status)
		 VALUES ($1, $2, $3, 'pending')
		 RETURNING id, type, targets, domain, status, created_at, completed_at`,
		taskType, req.Targets, req.Domain,
	).Scan(&task.ID, &task.Type, &task.Targets, &task.Domain, &task.Status, &task.CreatedAt, &task.CompletedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task: " + err.Error()})
		return
	}

	// TODO: Phase 3 — dispatch to edge nodes via gRPC

	c.JSON(http.StatusAccepted, task)
}
