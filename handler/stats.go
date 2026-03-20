package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatsHandler struct {
	db *pgxpool.Pool
}

func NewStatsHandler(db *pgxpool.Pool) *StatsHandler {
	return &StatsHandler{db: db}
}

type StatsOverview struct {
	TotalDomains  int     `json:"total_domains"`
	ActiveNodes   int     `json:"active_nodes"`
	TotalRequests int64   `json:"total_requests"`
	CacheHitRate  float64 `json:"cache_hit_rate"`
}

func (h *StatsHandler) Overview(c *gin.Context) {
	var totalDomains int
	err := h.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM domains").Scan(&totalDomains)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query domains: " + err.Error()})
		return
	}

	var activeNodes int
	_ = h.db.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM nodes WHERE status = 'online'").Scan(&activeNodes)

	// Placeholder stats - will connect to Prometheus later
	overview := StatsOverview{
		TotalDomains:  totalDomains,
		ActiveNodes:   activeNodes,
		TotalRequests: 0,
		CacheHitRate:  0.0,
	}

	c.JSON(http.StatusOK, overview)
}

type BandwidthPoint struct {
	Timestamp time.Time `json:"timestamp"`
	BytesIn   int64     `json:"bytes_in"`
	BytesOut  int64     `json:"bytes_out"`
}

type BandwidthResp struct {
	Domain string           `json:"domain"`
	From   string           `json:"from"`
	To     string           `json:"to"`
	Data   []BandwidthPoint `json:"data"`
}

func (h *StatsHandler) Bandwidth(c *gin.Context) {
	domain := c.Query("domain")
	from := c.Query("from")
	to := c.Query("to")

	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain query parameter is required"})
		return
	}
	if from == "" {
		from = time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	}
	if to == "" {
		to = time.Now().Format(time.RFC3339)
	}

	// Placeholder data - will connect to Prometheus later
	resp := BandwidthResp{
		Domain: domain,
		From:   from,
		To:     to,
		Data:   []BandwidthPoint{},
	}

	c.JSON(http.StatusOK, resp)
}
