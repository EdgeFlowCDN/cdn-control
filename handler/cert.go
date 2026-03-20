package handler

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EdgeFlowCDN/cdn-control/model"
)

type CertHandler struct {
	db *pgxpool.Pool
}

func NewCertHandler(db *pgxpool.Pool) *CertHandler {
	return &CertHandler{db: db}
}

func (h *CertHandler) Upload(c *gin.Context) {
	domainID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}

	var req model.UploadCertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse certificate to extract metadata
	block, _ := pem.Decode([]byte(req.CertPEM))
	if block == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid certificate PEM"})
		return
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid certificate: " + err.Error()})
		return
	}

	var result model.Certificate
	err = h.db.QueryRow(context.Background(),
		`INSERT INTO certificates (domain_id, cert_pem, key_pem, issuer, not_before, not_after, auto_renew)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, domain_id, issuer, not_before, not_after, auto_renew, created_at`,
		domainID, req.CertPEM, req.KeyPEM, cert.Issuer.CommonName,
		cert.NotBefore, cert.NotAfter, req.AutoRenew,
	).Scan(&result.ID, &result.DomainID, &result.Issuer, &result.NotBefore, &result.NotAfter, &result.AutoRenew, &result.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store certificate: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *CertHandler) List(c *gin.Context) {
	rows, err := h.db.Query(context.Background(),
		`SELECT id, domain_id, issuer, not_before, not_after, auto_renew, created_at
		 FROM certificates ORDER BY id`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var certs []model.Certificate
	for rows.Next() {
		var cert model.Certificate
		if err := rows.Scan(&cert.ID, &cert.DomainID, &cert.Issuer, &cert.NotBefore, &cert.NotAfter, &cert.AutoRenew, &cert.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		certs = append(certs, cert)
	}
	if certs == nil {
		certs = []model.Certificate{}
	}

	c.JSON(http.StatusOK, certs)
}

func (h *CertHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cert id"})
		return
	}

	tag, err := h.db.Exec(context.Background(), "DELETE FROM certificates WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "certificate not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
