package grpc

// This file defines the gRPC service types manually without protobuf code generation.
// For production, these should be generated from .proto files.

import (
	"context"

	"google.golang.org/grpc"
)

// NodeInfo identifies an edge node.
type NodeInfo struct {
	NodeId        string `json:"node_id"`
	Ip            string `json:"ip"`
	ConfigVersion int64  `json:"config_version"`
}

type FullConfigResponse struct {
	Version    int64  `json:"version"`
	ConfigJson string `json:"config_json"`
}

type ConfigUpdateResponse struct {
	UpdateJson string `json:"update_json"`
	Version    int64  `json:"version"`
}

type PurgeRequest struct {
	TaskId  string   `json:"task_id"`
	Type    string   `json:"type"`
	Targets []string `json:"targets"`
	Domain  string   `json:"domain"`
}

type PurgeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type HeartbeatRequest struct {
	NodeId       string  `json:"node_id"`
	Timestamp    int64   `json:"timestamp"`
	CpuUsage     float64 `json:"cpu_usage"`
	MemUsage     float64 `json:"mem_usage"`
	BandwidthBps int64   `json:"bandwidth_bps"`
	Connections  int64   `json:"connections"`
	DiskUsed     int64   `json:"disk_used"`
	DiskTotal    int64   `json:"disk_total"`
}

type HeartbeatResponse struct {
	Ok bool `json:"ok"`
}

// EdgeServiceServer is the server API for EdgeService.
type EdgeServiceServer interface {
	GetFullConfig(context.Context, *NodeInfo) (*FullConfigResponse, error)
	WatchConfig(*NodeInfo, EdgeService_WatchConfigServer) error
	PurgeNotify(context.Context, *PurgeRequest) (*PurgeResponse, error)
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
}

// UnimplementedEdgeServiceServer provides default implementations.
type UnimplementedEdgeServiceServer struct{}

func (UnimplementedEdgeServiceServer) GetFullConfig(context.Context, *NodeInfo) (*FullConfigResponse, error) {
	return nil, grpc.ErrServerStopped
}
func (UnimplementedEdgeServiceServer) WatchConfig(*NodeInfo, EdgeService_WatchConfigServer) error {
	return grpc.ErrServerStopped
}
func (UnimplementedEdgeServiceServer) PurgeNotify(context.Context, *PurgeRequest) (*PurgeResponse, error) {
	return nil, grpc.ErrServerStopped
}
func (UnimplementedEdgeServiceServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, grpc.ErrServerStopped
}

// EdgeService_WatchConfigServer is the server-side streaming interface.
type EdgeService_WatchConfigServer interface {
	Send(*ConfigUpdateResponse) error
	grpc.ServerStream
}

type edgeServiceWatchConfigServer struct {
	grpc.ServerStream
}

func (s *edgeServiceWatchConfigServer) Send(resp *ConfigUpdateResponse) error {
	return s.ServerStream.SendMsg(resp)
}

// Service description for registration
var edgeServiceDesc = grpc.ServiceDesc{
	ServiceName: "edgeflow.EdgeService",
	HandlerType: (*EdgeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "GetFullConfig", Handler: getFullConfigHandler},
		{MethodName: "PurgeNotify", Handler: purgeNotifyHandler},
		{MethodName: "Heartbeat", Handler: heartbeatHandler},
	},
	Streams: []grpc.StreamDesc{
		{StreamName: "WatchConfig", Handler: watchConfigHandler, ServerStreams: true},
	},
}

func RegisterEdgeServiceServer(s *grpc.Server, srv EdgeServiceServer) {
	s.RegisterService(&edgeServiceDesc, srv)
}

func getFullConfigHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(EdgeServiceServer).GetFullConfig(ctx, in)
}

func purgeNotifyHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PurgeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(EdgeServiceServer).PurgeNotify(ctx, in)
}

func heartbeatHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	return srv.(EdgeServiceServer).Heartbeat(ctx, in)
}

func watchConfigHandler(srv interface{}, stream grpc.ServerStream) error {
	in := new(NodeInfo)
	if err := stream.RecvMsg(in); err != nil {
		return err
	}
	return srv.(EdgeServiceServer).WatchConfig(in, &edgeServiceWatchConfigServer{stream})
}
