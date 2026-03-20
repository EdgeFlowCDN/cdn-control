package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EdgeFlowCDN/cdn-control/db"
	"github.com/EdgeFlowCDN/cdn-control/handler"
	"github.com/EdgeFlowCDN/cdn-control/middleware"
)

var (
	testDB     *pgxpool.Pool
	testRouter *gin.Engine
	testToken  string
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		fmt.Println("SKIP: TEST_DATABASE_URL not set, skipping integration tests")
		os.Exit(0)
	}

	var err error
	testDB, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		fmt.Printf("failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer testDB.Close()

	if err := db.Migrate(testDB); err != nil {
		fmt.Printf("failed to migrate: %v\n", err)
		os.Exit(1)
	}

	// Clean tables before tests
	cleanDB()

	middleware.SetJWTSecret("test-secret")
	testToken, _ = middleware.GenerateToken(1, "admin", "admin", 24)
	testRouter = handler.SetupRouter(testDB, 24)

	code := m.Run()

	cleanDB()
	os.Exit(code)
}

func cleanDB() {
	tables := []string{"purge_tasks", "cache_rules", "certificates", "origins", "domains", "nodes", "users"}
	for _, t := range tables {
		testDB.Exec(context.Background(), "DELETE FROM "+t)
	}
}

func doRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

func TestDomainCRUD(t *testing.T) {
	// Create
	w := doRequest("POST", "/api/v1/domains", map[string]string{
		"domain": "test.example.com",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create domain: status = %d, body = %s", w.Code, w.Body.String())
	}

	var domain struct {
		ID     int64  `json:"id"`
		Domain string `json:"domain"`
		Status string `json:"status"`
	}
	json.Unmarshal(w.Body.Bytes(), &domain)
	if domain.Domain != "test.example.com" {
		t.Errorf("domain = %q, want test.example.com", domain.Domain)
	}
	if domain.Status != "active" {
		t.Errorf("status = %q, want active", domain.Status)
	}

	// List
	w = doRequest("GET", "/api/v1/domains", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list domains: status = %d", w.Code)
	}

	var domains []struct{ ID int64 }
	json.Unmarshal(w.Body.Bytes(), &domains)
	if len(domains) < 1 {
		t.Error("expected at least 1 domain")
	}

	// Get
	w = doRequest("GET", fmt.Sprintf("/api/v1/domains/%d", domain.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get domain: status = %d", w.Code)
	}

	// Update
	w = doRequest("PUT", fmt.Sprintf("/api/v1/domains/%d", domain.ID), map[string]string{
		"status": "disabled",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("update domain: status = %d, body = %s", w.Code, w.Body.String())
	}

	// Delete
	w = doRequest("DELETE", fmt.Sprintf("/api/v1/domains/%d", domain.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete domain: status = %d", w.Code)
	}

	// Verify deleted
	w = doRequest("GET", fmt.Sprintf("/api/v1/domains/%d", domain.ID), nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("get deleted domain: status = %d, want 404", w.Code)
	}
}

func TestOriginCRUD(t *testing.T) {
	// Create domain first
	w := doRequest("POST", "/api/v1/domains", map[string]string{
		"domain": "origin-test.example.com",
	})
	var domain struct{ ID int64 }
	json.Unmarshal(w.Body.Bytes(), &domain)

	// Create origin
	w = doRequest("POST", fmt.Sprintf("/api/v1/domains/%d/origins", domain.ID), map[string]interface{}{
		"addr":   "https://origin.example.com",
		"weight": 100,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create origin: status = %d, body = %s", w.Code, w.Body.String())
	}

	var origin struct{ ID int64 }
	json.Unmarshal(w.Body.Bytes(), &origin)

	// List origins
	w = doRequest("GET", fmt.Sprintf("/api/v1/domains/%d/origins", domain.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list origins: status = %d", w.Code)
	}

	// Delete origin
	w = doRequest("DELETE", fmt.Sprintf("/api/v1/domains/%d/origins/%d", domain.ID, origin.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete origin: status = %d", w.Code)
	}

	// Cleanup
	doRequest("DELETE", fmt.Sprintf("/api/v1/domains/%d", domain.ID), nil)
}

func TestPurgeTask(t *testing.T) {
	// Create purge task
	w := doRequest("POST", "/api/v1/purge/url", map[string]interface{}{
		"targets": []string{"https://test.example.com/image.png"},
		"domain":  "test.example.com",
	})
	if w.Code != http.StatusAccepted {
		t.Fatalf("create purge: status = %d, body = %s", w.Code, w.Body.String())
	}

	var task struct {
		ID     int64  `json:"id"`
		Type   string `json:"type"`
		Status string `json:"status"`
	}
	json.Unmarshal(w.Body.Bytes(), &task)
	if task.Type != "url" {
		t.Errorf("type = %q, want url", task.Type)
	}
	if task.Status != "pending" {
		t.Errorf("status = %q, want pending", task.Status)
	}

	// Get task
	w = doRequest("GET", fmt.Sprintf("/api/v1/purge/tasks/%d", task.ID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get task: status = %d", w.Code)
	}
}

func TestAuthRequired(t *testing.T) {
	// Request without token
	req, _ := http.NewRequest("GET", "/api/v1/domains", nil)
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("no auth: status = %d, want 401", w.Code)
	}
}
