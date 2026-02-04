package server

import (
	"context"
	"time"

	"kvstore/pb"
	"kvstore/store"
)

type GRPCServer struct {
	pb.UnimplementedKVStoreServer
	store *store.Store
}

func NewGRPCServer(s *store.Store) *GRPCServer {
	return &GRPCServer{store: s}
}

func (g *GRPCServer) Get(_ context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	val, ok := g.store.Get(req.Key)
	return &pb.GetResponse{Value: val, Found: ok}, nil
}

func (g *GRPCServer) Set(_ context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	var ttl time.Duration
	if req.TtlSeconds > 0 {
		ttl = time.Duration(req.TtlSeconds) * time.Second
	}
	g.store.Set(req.Key, req.Value, ttl)
	return &pb.SetResponse{}, nil
}

func (g *GRPCServer) Delete(_ context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	deleted := g.store.Delete(req.Key)
	return &pb.DeleteResponse{Deleted: deleted}, nil
}
