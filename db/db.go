package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect creates a PostgreSQL connection pool.
func Connect(dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// Migrate runs database migrations.
func Migrate(pool *pgxpool.Pool) error {
	ctx := context.Background()
	_, err := pool.Exec(ctx, migrationSQL)
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

const migrationSQL = `
CREATE TABLE IF NOT EXISTS domains (
    id          BIGSERIAL PRIMARY KEY,
    domain      VARCHAR(255) NOT NULL UNIQUE,
    cname       VARCHAR(255) NOT NULL DEFAULT '',
    status      VARCHAR(20) DEFAULT 'pending',
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS origins (
    id          BIGSERIAL PRIMARY KEY,
    domain_id   BIGINT NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    addr        VARCHAR(500) NOT NULL,
    port        INT DEFAULT 443,
    weight      INT DEFAULT 100,
    priority    INT DEFAULT 0,
    protocol    VARCHAR(10) DEFAULT 'https',
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cache_rules (
    id            BIGSERIAL PRIMARY KEY,
    domain_id     BIGINT NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    path_pattern  VARCHAR(500) DEFAULT '/*',
    ttl           INT NOT NULL,
    ignore_query  BOOLEAN DEFAULT FALSE,
    priority      INT DEFAULT 0,
    created_at    TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS certificates (
    id            BIGSERIAL PRIMARY KEY,
    domain_id     BIGINT NOT NULL REFERENCES domains(id) ON DELETE CASCADE,
    cert_pem      TEXT NOT NULL,
    key_pem       TEXT NOT NULL,
    issuer        VARCHAR(100) DEFAULT '',
    not_before    TIMESTAMP,
    not_after     TIMESTAMP,
    auto_renew    BOOLEAN DEFAULT FALSE,
    created_at    TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nodes (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL UNIQUE,
    ip              VARCHAR(45) NOT NULL,
    region          VARCHAR(50) DEFAULT '',
    isp             VARCHAR(20) DEFAULT '',
    status          VARCHAR(20) DEFAULT 'offline',
    max_bandwidth   BIGINT DEFAULT 0,
    last_heartbeat  TIMESTAMP,
    created_at      TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS purge_tasks (
    id          BIGSERIAL PRIMARY KEY,
    type        VARCHAR(10) NOT NULL,
    targets     TEXT[] NOT NULL,
    domain      VARCHAR(255) NOT NULL,
    status      VARCHAR(20) DEFAULT 'pending',
    created_at  TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(100) NOT NULL UNIQUE,
    password    VARCHAR(255) NOT NULL,
    role        VARCHAR(20) DEFAULT 'user',
    created_at  TIMESTAMP DEFAULT NOW()
);
`
