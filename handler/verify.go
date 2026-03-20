package handler

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VerifyHandler struct {
	db *pgxpool.Pool
}

func NewVerifyHandler(db *pgxpool.Pool) *VerifyHandler {
	return &VerifyHandler{db: db}
}

type VerifyResp struct {
	DomainID       int64  `json:"domain_id"`
	Domain         string `json:"domain"`
	ExpectedCNAME  string `json:"expected_cname"`
	ActualCNAME    string `json:"actual_cname"`
	Verified       bool   `json:"verified"`
	Status         string `json:"status"`
}

func (h *VerifyHandler) Verify(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var domain, cname string
	err = h.db.QueryRow(context.Background(),
		"SELECT domain, cname FROM domains WHERE id = $1", id,
	).Scan(&domain, &cname)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		return
	}

	// Look up the CNAME record for the domain
	actual, err := net.LookupCNAME(domain)
	if err != nil {
		// DNS lookup failed - domain not yet configured
		c.JSON(http.StatusOK, VerifyResp{
			DomainID:      id,
			Domain:        domain,
			ExpectedCNAME: cname,
			ActualCNAME:   "",
			Verified:      false,
			Status:        "pending",
		})
		return
	}

	// Normalize: remove trailing dots for comparison
	actualNorm := strings.TrimSuffix(actual, ".")
	expectedNorm := strings.TrimSuffix(cname, ".")

	verified := strings.EqualFold(actualNorm, expectedNorm)

	newStatus := "pending"
	if verified {
		newStatus = "verified"
		_, _ = h.db.Exec(context.Background(),
			"UPDATE domains SET status = 'verified', updated_at = NOW() WHERE id = $1", id)
	}

	c.JSON(http.StatusOK, VerifyResp{
		DomainID:      id,
		Domain:        domain,
		ExpectedCNAME: cname,
		ActualCNAME:   actualNorm,
		Verified:      verified,
		Status:        newStatus,
	})
}
