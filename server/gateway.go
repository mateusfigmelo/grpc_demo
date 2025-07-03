package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	pb "example/grpc_demo/library"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func StartGateway() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a more basic ServeMux without custom header matching
	mux := runtime.NewServeMux()

	// Use basic connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	err := pb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
	if err != nil {
		log.Fatalf("Failed to register UserService gateway: %v", err)
	}

	err = pb.RegisterLibraryServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
	if err != nil {
		log.Fatalf("Failed to register LibraryService gateway: %v", err)
	}

	// Add CORS middleware
	handler := corsMiddleware(mux)

	fmt.Println("REST Gateway server starting on port 8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Failed to serve gateway: %v", err)
	}
}

func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Gateway request: %s %s\n", r.Method, r.URL.Path)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}
