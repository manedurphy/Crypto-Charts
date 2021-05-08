package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/manedurphy/grpc-web/pb"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("could not start gateway: %v", err)
	}
}

func run() error {
	fmt.Println("starting grpc gateway on port 8081")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mux := runtime.NewServeMux()
	conn, err := grpc.DialContext(ctx, os.Getenv("server-url"), grpc.WithInsecure()) // Try to use TLS

	if err != nil {
		log.Fatalf("failed to dial gRPC server: %v", err)
	}
	err = pb.RegisterBitcoinServiceHandler(ctx, mux, conn)

	if err != nil {
		log.Fatalf("could not dial endpoint: %v", err)
	}

	goEnv := os.Getenv("GO_ENV")

	err = healthHandler(mux)

	if err != nil {
		log.Fatalf("gateway is unhealthy: %v", err)
	}

	if goEnv == "development" {
		handler := cors.Default().Handler(mux)
		return http.ListenAndServe(":8081", handler)
	}

	return http.ListenAndServe(":8081", mux)
}

func healthHandler(mux *runtime.ServeMux) error {
	return mux.HandlePath("GET", "/healthz", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		fmt.Println("gateway is healthy!")
		w.Write([]byte("healthy"))
	})
}
