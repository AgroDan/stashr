package main

import (
	"context"
	"flag"
	"fmt"
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

	// By default, this application will start an HTTP server on port 8080 and a gRPC server on port 9090.
	// However, with the appropriate flags, you can disable the HTTP server, gRPC server, or change the
	// port to an arbitrary number.
	httpPort := flag.Int("hport", 8080, "HTTP Port to listen on.")
	grpcPort := flag.Int("gport", 9090, "gRPC Port to listen on.")
	disableHttp := flag.Bool("disableHTTP", false, "Disable HTTP Service")
	disablegRPC := flag.Bool("disableGRPC", false, "Disable gRPC Service")

	flag.Parse()

	// HTTP server
	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *httpPort),
		Handler: server.NewHTTPServer(s).Handler(),
	}

	// gRPC server
	grpcSrv := grpc.NewServer()
	pb.RegisterKVStoreServer(grpcSrv, server.NewGRPCServer(s))
	reflection.Register(grpcSrv)

	// Start HTTP
	if !*disableHttp {
		go func() {
			log.Printf("HTTP server listening on :%d\n", *httpPort)
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("HTTP server error: %v", err)
			}
		}()
	}

	// Start gRPC
	if !*disablegRPC {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
		if err != nil {
			log.Fatalf("failed to listen on :%d: %v", *grpcPort, err)
		}
		go func() {
			log.Printf("gRPC server listening on :%d\n", *grpcPort)
			if err := grpcSrv.Serve(lis); err != nil {
				log.Fatalf("gRPC server error: %v", err)
			}
		}()
	}

	if *disableHttp && *disablegRPC {
		log.Fatalf("All servers disabled! What should I do?")
	}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("shutting down...")

	if !*disablegRPC {
		grpcSrv.GracefulStop()
	}

	if !*disableHttp {
		httpSrv.Shutdown(context.Background())
	}
}
