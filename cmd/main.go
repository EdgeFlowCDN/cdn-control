package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EdgeFlowCDN/cdn-control/config"
	"github.com/EdgeFlowCDN/cdn-control/db"
	cdngrpc "github.com/EdgeFlowCDN/cdn-control/grpc"
	"github.com/EdgeFlowCDN/cdn-control/handler"
	"github.com/EdgeFlowCDN/cdn-control/middleware"
)

func main() {
	configPath := flag.String("config", "configs/control-config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Connect to database
	pool, err := db.Connect(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Run migrations
	if err := db.Migrate(pool); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Set JWT secret
	middleware.SetJWTSecret(cfg.JWT.Secret)

	// Create default admin user
	authH := handler.NewAuthHandler(pool, cfg.JWT.ExpireHour)
	if err := authH.InitAdmin(); err != nil {
		log.Printf("warning: failed to init admin user: %v", err)
	}

	// Start gRPC server
	if cfg.Server.GRPCListen != "" {
		grpcServer := cdngrpc.NewServer(pool, cfg.Server.GRPCListen)
		go func() {
			if err := grpcServer.Start(); err != nil {
				log.Fatalf("gRPC server failed: %v", err)
			}
		}()
	}

	// Setup router
	router := handler.SetupRouter(pool, cfg.JWT.ExpireHour)

	// Create HTTP server
	srv := &http.Server{
		Addr:    cfg.Server.Listen,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("EdgeFlow Control Plane starting on %s", cfg.Server.Listen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("received shutdown signal, shutting down gracefully...")

	// Graceful shutdown with 15s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	} else {
		log.Println("HTTP server stopped gracefully")
	}

	// Close DB pool after shutdown
	pool.Close()
	log.Println("database connection closed")
}
