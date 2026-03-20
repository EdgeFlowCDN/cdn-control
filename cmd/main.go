package main

import (
	"flag"
	"fmt"
	"log"
	"os"

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
	defer pool.Close()

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

	log.Printf("EdgeFlow Control Plane starting on %s", cfg.Server.Listen)
	if err := router.Run(cfg.Server.Listen); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
