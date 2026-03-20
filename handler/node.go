package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EdgeFlowCDN/cdn-control/model"
)

type NodeHandler struct {
	db *pgxpool.Pool
}

func NewNodeHandler(db *pgxpool.Pool) *NodeHandler {
	return &NodeHandler{db: db}
}

func (h *NodeHandler) List(c *gin.Context) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, name, ip, region, isp, status, max_bandwidth, last_heartbeat, created_at
		 FROM nodes ORDER BY id`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var nodes []model.Node
	for rows.Next() {
		var n model.Node
		if err := rows.Scan(&n.ID, &n.Name, &n.IP, &n.Region, &n.ISP, &n.Status, &n.MaxBandwidth, &n.LastHeartbeat, &n.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		nodes = append(nodes, n)
	}
	if nodes == nil {
		nodes = []model.Node{}
	}

	c.JSON(http.StatusOK, nodes)
}

func (h *NodeHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var n model.Node
	err = h.db.QueryRow(context.Background(),
		`SELECT id, name, ip, region, isp, status, max_bandwidth, last_heartbeat, created_at
		 FROM nodes WHERE id = $1`, id,
	).Scan(&n.ID, &n.Name, &n.IP, &n.Region, &n.ISP, &n.Status, &n.MaxBandwidth, &n.LastHeartbeat, &n.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	c.JSON(http.StatusOK, n)
}

func (h *NodeHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	valid := map[string]bool{"online": true, "offline": true, "maintenance": true}
	if !valid[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be online, offline, or maintenance"})
		return
	}

	tag, err := h.db.Exec(context.Background(),
		"UPDATE nodes SET status = $2 WHERE id = $1", id, req.Status,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}
