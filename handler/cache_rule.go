package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EdgeFlowCDN/cdn-control/model"
)

type CacheRuleHandler struct {
	db *pgxpool.Pool
}

func NewCacheRuleHandler(db *pgxpool.Pool) *CacheRuleHandler {
	return &CacheRuleHandler{db: db}
}

func (h *CacheRuleHandler) Create(c *gin.Context) {
	domainID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}

	var req model.CreateCacheRuleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PathPattern == "" {
		req.PathPattern = "/*"
	}

	var rule model.CacheRule
	err = h.db.QueryRow(context.Background(),
		`INSERT INTO cache_rules (domain_id, path_pattern, ttl, ignore_query, priority)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, domain_id, path_pattern, ttl, ignore_query, priority, created_at`,
		domainID, req.PathPattern, req.TTL, req.IgnoreQuery, req.Priority,
	).Scan(&rule.ID, &rule.DomainID, &rule.PathPattern, &rule.TTL, &rule.IgnoreQuery, &rule.Priority, &rule.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create cache rule: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

func (h *CacheRuleHandler) List(c *gin.Context) {
	domainID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}

	rows, err := h.db.Query(context.Background(),
		`SELECT id, domain_id, path_pattern, ttl, ignore_query, priority, created_at
		 FROM cache_rules WHERE domain_id = $1 ORDER BY priority DESC, id`,
		domainID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var rules []model.CacheRule
	for rows.Next() {
		var r model.CacheRule
		if err := rows.Scan(&r.ID, &r.DomainID, &r.PathPattern, &r.TTL, &r.IgnoreQuery, &r.Priority, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rules = append(rules, r)
	}
	if rules == nil {
		rules = []model.CacheRule{}
	}

	c.JSON(http.StatusOK, rules)
}

func (h *CacheRuleHandler) Delete(c *gin.Context) {
	ruleID, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}

	tag, err := h.db.Exec(context.Background(), "DELETE FROM cache_rules WHERE id = $1", ruleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "cache rule not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
