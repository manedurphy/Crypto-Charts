package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/manedurphy/grpc-web/pb"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	secure = flag.Bool("secure", true, "set to true to use TLS connection")
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("could not start gateway: %v", err)
	}
}

func run() error {
	fmt.Println("starting grpc gateway on port 8081")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	mux := runtime.NewServeMux()

	creds, _ := credentials.NewClientTLSFromFile("tls/ca.crt", "")

	var conn *grpc.ClientConn
	var err error
	if !*secure {
		fmt.Println("Insecure connection established with server")
		conn, err = grpc.DialContext(ctx, os.Getenv("BTC_SERVER"), grpc.WithInsecure())
	} else {
		fmt.Println("TLS connection established with server")
		conn, err = grpc.DialContext(ctx, os.Getenv("BTC_SERVER"), grpc.WithTransportCredentials(creds))
	}

	if err != nil {
		log.Fatalf("failed to dial gRPC server: %v", err)
	}
	// err = pb.RegisterBitcoinServiceHandler(ctx, mux, conn)

	// if err != nil {
	// 	log.Fatalf("could not dial endpoint: %v", err)
	// }

	err = pb.RegisterCryptoServiceHandler(ctx, mux, conn)

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
