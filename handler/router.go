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

	// Public routes
	r.POST("/api/v1/auth/login", authH.Login)

	// Protected routes
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth())
	{
		// Domains
		api.POST("/domains", domainH.Create)
		api.GET("/domains", domainH.List)
		api.GET("/domains/:id", domainH.Get)
		api.PUT("/domains/:id", domainH.Update)
		api.DELETE("/domains/:id", domainH.Delete)

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
	}

	return r
}
