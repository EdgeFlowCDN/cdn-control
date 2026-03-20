package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EdgeFlowCDN/cdn-control/middleware"
)

// SetupRouter creates and configures the Gin router with all API routes.
func SetupRouter(db *pgxpool.Pool, expireHours int) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	domainH := NewDomainHandler(db)
	originH := NewOriginHandler(db)
	cacheRuleH := NewCacheRuleHandler(db)
	purgeH := NewPurgeHandler(db)
	certH := NewCertHandler(db)
	nodeH := NewNodeHandler(db)
	authH := NewAuthHandler(db, expireHours)
	statsH := NewStatsHandler(db)
	webhookH := NewWebhookHandler(db)
	userH := NewUserHandler(db)
	auditH := NewAuditHandler(db)
	verifyH := NewVerifyHandler(db)

	// Service info and health
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "EdgeFlow Control Plane",
			"version": "1.0.0",
			"docs":    "/swagger/",
			"api":     "/api/v1",
		})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	r.POST("/api/v1/auth/login", authH.Login)

	// Protected routes
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth())
	api.Use(middleware.AuditLog(db))
	{
		// Domains
		api.POST("/domains", domainH.Create)
		api.GET("/domains", domainH.List)
		api.GET("/domains/:id", domainH.Get)
		api.PUT("/domains/:id", domainH.Update)
		api.DELETE("/domains/:id", domainH.Delete)

		// Domain CNAME verification
		api.POST("/domains/:id/verify", verifyH.Verify)

		// Origins (sub-resource of domain)
		api.POST("/domains/:id/origins", originH.Create)
		api.GET("/domains/:id/origins", originH.List)
		api.PUT("/domains/:id/origins/:oid", originH.Update)
		api.DELETE("/domains/:id/origins/:oid", originH.Delete)

		// Cache rules (sub-resource of domain)
		api.POST("/domains/:id/cache-rules", cacheRuleH.Create)
		api.GET("/domains/:id/cache-rules", cacheRuleH.List)
		api.DELETE("/domains/:id/cache-rules/:rid", cacheRuleH.Delete)

		// Cache operations
		api.POST("/purge/url", purgeH.PurgeURL)
		api.POST("/purge/dir", purgeH.PurgeDir)
		api.POST("/purge/all", purgeH.PurgeAll)
		api.POST("/prefetch", purgeH.Prefetch)
		api.GET("/purge/tasks/:id", purgeH.GetTask)

		// Certificates
		api.POST("/domains/:id/certs", certH.Upload)
		api.GET("/certs", certH.List)
		api.DELETE("/certs/:cid", certH.Delete)

		// Nodes
		api.GET("/nodes", nodeH.List)
		api.GET("/nodes/:id", nodeH.Get)
		api.PUT("/nodes/:id/status", nodeH.UpdateStatus)

		// Stats
		api.GET("/stats/overview", statsH.Overview)
		api.GET("/stats/bandwidth", statsH.Bandwidth)

		// Webhooks
		api.POST("/webhooks", webhookH.Create)
		api.GET("/webhooks", webhookH.List)
		api.DELETE("/webhooks/:id", webhookH.Delete)

		// Users - current user
		api.GET("/users/me", userH.Me)
		api.PUT("/users/me/password", userH.ChangePassword)

		// Users - admin only
		admin := api.Group("")
		admin.Use(middleware.AdminOnly())
		{
			admin.POST("/users", userH.Create)
			admin.GET("/users", userH.List)
		}

		// Audit logs
		api.GET("/audit-logs", auditH.List)
	}

	// Swagger / OpenAPI documentation (no auth required)
	RegisterSwaggerRoutes(r)

	return r
}
