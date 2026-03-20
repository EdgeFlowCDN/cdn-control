package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGenerateAndValidateToken(t *testing.T) {
	SetJWTSecret("test-secret-key")

	token, err := GenerateToken(1, "admin", "admin", 24)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
}

func TestJWTAuthMiddleware(t *testing.T) {
	SetJWTSecret("test-secret-key")

	router := gin.New()
	router.Use(JWTAuth())
	router.GET("/test", func(c *gin.Context) {
		username, _ := c.Get("username")
		c.JSON(200, gin.H{"username": username})
	})

	// No token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("no token: status = %d, want 401", w.Code)
	}

	// Invalid token
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("invalid token: status = %d, want 401", w.Code)
	}

	// Valid token
	token, _ := GenerateToken(1, "testuser", "admin", 24)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("valid token: status = %d, want 200", w.Code)
	}
}

func TestAdminOnly(t *testing.T) {
	SetJWTSecret("test-secret-key")

	router := gin.New()
	router.Use(JWTAuth(), AdminOnly())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	// Admin user
	token, _ := GenerateToken(1, "admin", "admin", 24)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("admin: status = %d, want 200", w.Code)
	}

	// Regular user
	token, _ = GenerateToken(2, "user", "user", 24)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("regular user: status = %d, want 403", w.Code)
	}
}

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("mypassword")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword("mypassword", hash) {
		t.Error("correct password should match")
	}
	if CheckPassword("wrongpassword", hash) {
		t.Error("wrong password should not match")
	}
}
