package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EdgeFlowCDN/cdn-control/model"
)

type DomainHandler struct {
	db *pgxpool.Pool
}

func NewDomainHandler(db *pgxpool.Pool) *DomainHandler {
	return &DomainHandler{db: db}
}

func (h *DomainHandler) Create(c *gin.Context) {
	var req model.CreateDomainReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.CNAME == "" {
		req.CNAME = fmt.Sprintf("%s.edgeflow.dev", req.Domain)
	}

	var domain model.Domain
	err := h.db.QueryRow(context.Background(),
		`INSERT INTO domains (domain, cname, status) VALUES ($1, $2, 'active')
		 RETURNING id, domain, cname, status, created_at, updated_at`,
		req.Domain, req.CNAME,
	).Scan(&domain.ID, &domain.Domain, &domain.CNAME, &domain.Status, &domain.CreatedAt, &domain.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create domain: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, domain)
	notifyConfigChange()
}

func (h *DomainHandler) List(c *gin.Context) {
	rows, err := h.db.Query(context.Background(),
		"SELECT id, domain, cname, status, created_at, updated_at FROM domains ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var domains []model.Domain
	for rows.Next() {
		var d model.Domain
		if err := rows.Scan(&d.ID, &d.Domain, &d.CNAME, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		domains = append(domains, d)
	}
	if domains == nil {
		domains = []model.Domain{}
	}

	c.JSON(http.StatusOK, domains)
}

func (h *DomainHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var d model.Domain
	err = h.db.QueryRow(context.Background(),
		"SELECT id, domain, cname, status, created_at, updated_at FROM domains WHERE id = $1", id,
	).Scan(&d.ID, &d.Domain, &d.CNAME, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		return
	}

	c.JSON(http.StatusOK, d)
}

func (h *DomainHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req model.UpdateDomainReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var d model.Domain
	err = h.db.QueryRow(context.Background(),
		`UPDATE domains SET
			status = COALESCE(NULLIF($2, ''), status),
			cname = COALESCE(NULLIF($3, ''), cname),
			updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, domain, cname, status, created_at, updated_at`,
		id, req.Status, req.CNAME,
	).Scan(&d.ID, &d.Domain, &d.CNAME, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		return
	}

	c.JSON(http.StatusOK, d)
	notifyConfigChange()
}

func (h *DomainHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	tag, err := h.db.Exec(context.Background(), "DELETE FROM domains WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	notifyConfigChange()
}
