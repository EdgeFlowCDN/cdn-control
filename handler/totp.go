package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"

	"github.com/EdgeFlowCDN/cdn-control/middleware"
	"github.com/EdgeFlowCDN/cdn-control/model"
)

type TOTPHandler struct {
	db *pgxpool.Pool
}

func NewTOTPHandler(db *pgxpool.Pool) *TOTPHandler {
	return &TOTPHandler{db: db}
}

// Setup generates a new TOTP secret and stores it, but does not enable 2FA yet.
func (h *TOTPHandler) Setup(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "EdgeFlow CDN",
		AccountName: username.(string),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate TOTP secret"})
		return
	}

	_, err = h.db.Exec(context.Background(),
		"UPDATE users SET totp_secret = $1 WHERE id = $2",
		key.Secret(), userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save TOTP secret"})
		return
	}

	backupCodes := generateBackupCodes(8)

	c.JSON(http.StatusOK, model.Setup2FAResp{
		Secret:      key.Secret(),
		URL:         key.URL(),
		BackupCodes: backupCodes,
	})
}

// Verify validates a TOTP code and enables 2FA if correct.
func (h *TOTPHandler) Verify(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req model.Verify2FAReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var secret string
	err := h.db.QueryRow(context.Background(),
		"SELECT totp_secret FROM users WHERE id = $1",
		userID,
	).Scan(&secret)
	if err != nil || secret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA not set up yet"})
		return
	}

	if !totp.Validate(req.Code, secret) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid TOTP code"})
		return
	}

	_, err = h.db.Exec(context.Background(),
		"UPDATE users SET totp_enabled = TRUE WHERE id = $1",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enable 2FA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

// Disable disables 2FA after verifying the current TOTP code and password.
func (h *TOTPHandler) Disable(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req model.Disable2FAReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var secret string
	var passwordHash string
	err := h.db.QueryRow(context.Background(),
		"SELECT totp_secret, password FROM users WHERE id = $1",
		userID,
	).Scan(&secret, &passwordHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query user"})
		return
	}

	if !middleware.CheckPassword(req.Password, passwordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	if !totp.Validate(req.Code, secret) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid TOTP code"})
		return
	}

	_, err = h.db.Exec(context.Background(),
		"UPDATE users SET totp_enabled = FALSE, totp_secret = '' WHERE id = $1",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to disable 2FA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}

// Status returns whether 2FA is enabled for the current user.
func (h *TOTPHandler) Status(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var enabled bool
	err := h.db.QueryRow(context.Background(),
		"SELECT totp_enabled FROM users WHERE id = $1",
		userID,
	).Scan(&enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query 2FA status"})
		return
	}

	c.JSON(http.StatusOK, model.TwoFAStatusResp{Enabled: enabled})
}

func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		b := make([]byte, 4)
		_, _ = rand.Read(b)
		codes[i] = fmt.Sprintf("%s-%s", hex.EncodeToString(b[:2]), hex.EncodeToString(b[2:]))
	}
	return codes
}
