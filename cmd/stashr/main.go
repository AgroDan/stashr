package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"stashr/pb"
	"stashr/server"
	"stashr/store"
)

func main() {
	s := store.New()
	defer s.Stop()

	// HTTP server
	httpSrv := &http.Server{
		Addr:    ":8080",
		Handler: server.NewHTTPServer(s).Handler(),
	}

	// gRPC server
	grpcSrv := grpc.NewServer()
	pb.RegisterKVStoreServer(grpcSrv, server.NewGRPCServer(s))
	reflection.Register(grpcSrv)

	// Start HTTP
	go func() {
		log.Println("HTTP server listening on :8080")
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start gRPC
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen on :9090: %v", err)
	}
	go func() {
		log.Println("gRPC server listening on :9090")
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("shutting down...")
	grpcSrv.GracefulStop()
	httpSrv.Shutdown(context.Background())
}
