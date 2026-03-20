package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EdgeFlowCDN/cdn-control/model"
)

type OriginHandler struct {
	db *pgxpool.Pool
}

func NewOriginHandler(db *pgxpool.Pool) *OriginHandler {
	return &OriginHandler{db: db}
}

func (h *OriginHandler) Create(c *gin.Context) {
	domainID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}

	var req model.CreateOriginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Port == 0 {
		req.Port = 443
	}
	if req.Weight == 0 {
		req.Weight = 100
	}
	if req.Protocol == "" {
		req.Protocol = "https"
	}

	var o model.Origin
	err = h.db.QueryRow(context.Background(),
		`INSERT INTO origins (domain_id, addr, port, weight, priority, protocol)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, domain_id, addr, port, weight, priority, protocol, created_at`,
		domainID, req.Addr, req.Port, req.Weight, req.Priority, req.Protocol,
	).Scan(&o.ID, &o.DomainID, &o.Addr, &o.Port, &o.Weight, &o.Priority, &o.Protocol, &o.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create origin: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, o)
}

func (h *OriginHandler) List(c *gin.Context) {
	domainID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}

	rows, err := h.db.Query(context.Background(),
		"SELECT id, domain_id, addr, port, weight, priority, protocol, created_at FROM origins WHERE domain_id = $1 ORDER BY priority, id",
		domainID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var origins []model.Origin
	for rows.Next() {
		var o model.Origin
		if err := rows.Scan(&o.ID, &o.DomainID, &o.Addr, &o.Port, &o.Weight, &o.Priority, &o.Protocol, &o.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		origins = append(origins, o)
	}
	if origins == nil {
		origins = []model.Origin{}
	}

	c.JSON(http.StatusOK, origins)
}

func (h *OriginHandler) Update(c *gin.Context) {
	oid, err := strconv.ParseInt(c.Param("oid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid origin id"})
		return
	}

	var req model.UpdateOriginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var o model.Origin
	err = h.db.QueryRow(context.Background(),
		`UPDATE origins SET
			addr = COALESCE(NULLIF($2, ''), addr),
			port = CASE WHEN $3 > 0 THEN $3 ELSE port END,
			weight = CASE WHEN $4 > 0 THEN $4 ELSE weight END,
			priority = $5,
			protocol = COALESCE(NULLIF($6, ''), protocol)
		 WHERE id = $1
		 RETURNING id, domain_id, addr, port, weight, priority, protocol, created_at`,
		oid, req.Addr, req.Port, req.Weight, req.Priority, req.Protocol,
	).Scan(&o.ID, &o.DomainID, &o.Addr, &o.Port, &o.Weight, &o.Priority, &o.Protocol, &o.CreatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "origin not found"})
		return
	}

	c.JSON(http.StatusOK, o)
}

func (h *OriginHandler) Delete(c *gin.Context) {
	oid, err := strconv.ParseInt(c.Param("oid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid origin id"})
		return
	}

	tag, err := h.db.Exec(context.Background(), "DELETE FROM origins WHERE id = $1", oid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "origin not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
