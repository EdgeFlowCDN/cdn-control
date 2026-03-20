package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

// DomainConfig is the configuration pushed to edge nodes.
type DomainConfig struct {
	Host    string         `json:"host"`
	Origins []OriginConfig `json:"origins"`
	Cache   CacheConfig    `json:"cache"`
}

type OriginConfig struct {
	Addr     string `json:"addr"`
	Weight   int    `json:"weight"`
	Priority int    `json:"priority"`
}

type CacheConfig struct {
	DefaultTTL  string `json:"default_ttl"`
	IgnoreQuery bool   `json:"ignore_query"`
	ForceTTL    string `json:"force_ttl"`
}

// ConfigUpdate is sent to connected edge nodes.
type ConfigUpdate struct {
	Action  string        `json:"action"` // "full", "add", "update", "delete"
	Domains []DomainConfig `json:"domains,omitempty"`
	Domain  *DomainConfig  `json:"domain,omitempty"`
	Version int64          `json:"version"`
}

// PurgeCommand is sent to edge nodes to purge cache.
type PurgeCommand struct {
	TaskID  int64    `json:"task_id"`
	Type    string   `json:"type"` // "url", "dir", "all"
	Targets []string `json:"targets"`
	Domain  string   `json:"domain"`
}

// Server is the gRPC server for edge node communication.
type Server struct {
	UnimplementedEdgeServiceServer
	db         *pgxpool.Pool
	mu         sync.RWMutex
	edges      map[string]chan []byte // nodeID -> update channel
	version    int64
	listenAddr string
}

// NewServer creates a new gRPC server.
func NewServer(db *pgxpool.Pool, listenAddr string) *Server {
	return &Server{
		db:         db,
		edges:      make(map[string]chan []byte),
		version:    1,
		listenAddr: listenAddr,
	}
}

// Start starts the gRPC server.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	srv := grpc.NewServer()
	RegisterEdgeServiceServer(srv, s)
	log.Printf("[grpc] listening on %s", s.listenAddr)
	return srv.Serve(lis)
}

// GetFullConfig returns the complete domain configuration.
func (s *Server) GetFullConfig(ctx context.Context, req *NodeInfo) (*FullConfigResponse, error) {
	configs, err := s.loadDomainConfigs()
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(configs)
	return &FullConfigResponse{
		Version:    s.version,
		ConfigJson: string(data),
	}, nil
}

// WatchConfig streams configuration updates to edge nodes.
func (s *Server) WatchConfig(req *NodeInfo, stream EdgeService_WatchConfigServer) error {
	ch := make(chan []byte, 100)
	s.mu.Lock()
	s.edges[req.NodeId] = ch
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.edges, req.NodeId)
		s.mu.Unlock()
		log.Printf("[grpc] edge node %s disconnected", req.NodeId)
	}()

	log.Printf("[grpc] edge node %s connected (config version: %d)", req.NodeId, req.ConfigVersion)

	for {
		select {
		case data := <-ch:
			if err := stream.Send(&ConfigUpdateResponse{
				UpdateJson: string(data),
				Version:    s.version,
			}); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

// PurgeNotify sends a purge command to an edge node.
func (s *Server) PurgeNotify(ctx context.Context, req *PurgeRequest) (*PurgeResponse, error) {
	log.Printf("[grpc] purge request: type=%s domain=%s targets=%d", req.Type, req.Domain, len(req.Targets))
	return &PurgeResponse{
		Success: true,
		Message: "purge accepted",
	}, nil
}

// Heartbeat receives heartbeat from edge nodes.
func (s *Server) Heartbeat(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error) {
	log.Printf("[grpc] heartbeat from %s: cpu=%.1f%% mem=%.1f%% bw=%d conn=%d",
		req.NodeId, req.CpuUsage*100, req.MemUsage*100, req.BandwidthBps, req.Connections)
	return &HeartbeatResponse{Ok: true}, nil
}

// BroadcastUpdate sends a config update to all connected edge nodes.
func (s *Server) BroadcastUpdate(update *ConfigUpdate) {
	s.mu.Lock()
	s.version++
	update.Version = s.version
	s.mu.Unlock()

	data, err := json.Marshal(update)
	if err != nil {
		log.Printf("[grpc] marshal update error: %v", err)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for nodeID, ch := range s.edges {
		select {
		case ch <- data:
		default:
			log.Printf("[grpc] edge %s channel full, dropping update", nodeID)
		}
	}
}

// BroadcastPurge sends a purge command to all connected edge nodes.
func (s *Server) BroadcastPurge(cmd *PurgeCommand) {
	data, _ := json.Marshal(cmd)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.edges {
		select {
		case ch <- data:
		default:
		}
	}
}

func (s *Server) loadDomainConfigs() ([]DomainConfig, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT d.domain, o.addr, o.weight, o.priority
		 FROM domains d JOIN origins o ON d.id = o.domain_id
		 WHERE d.status = 'active' ORDER BY d.domain, o.priority`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	domainMap := make(map[string]*DomainConfig)
	var order []string
	for rows.Next() {
		var host, addr string
		var weight, priority int
		if err := rows.Scan(&host, &addr, &weight, &priority); err != nil {
			return nil, err
		}
		if _, exists := domainMap[host]; !exists {
			domainMap[host] = &DomainConfig{
				Host:  host,
				Cache: CacheConfig{DefaultTTL: "10m"},
			}
			order = append(order, host)
		}
		domainMap[host].Origins = append(domainMap[host].Origins, OriginConfig{
			Addr: addr, Weight: weight, Priority: priority,
		})
	}

	var configs []DomainConfig
	for _, host := range order {
		configs = append(configs, *domainMap[host])
	}
	return configs, nil
}
