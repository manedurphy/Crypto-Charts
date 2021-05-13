package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	pb "github.com/manedurphy/grpc-web/pb"
	"github.com/manedurphy/grpc-web/server/handlers"
	store "github.com/manedurphy/grpc-web/server/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	certfile = "tls/server.crt"
	keyfile  = "tls/server.key"

	secure = flag.Bool("secure", true, "set to true to use TLS connection")
)

type cryptoServer struct {
	pb.UnimplementedCryptoServiceServer
}

func main() {
	lis, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatalf("could not listen on port 8080: %v\n", err)
	}

	fmt.Println("gRPC server started on on port 8080")

	creds, err := credentials.NewServerTLSFromFile(certfile, keyfile)
	if err != nil {
		log.Fatalf("coudld not get certificates for tls: %v", err)
	}

	var s *grpc.Server
	if !*secure {
		fmt.Println("Insecure connection established with gateway")
		s = grpc.NewServer()
	} else {
		fmt.Println("TLS connection established with gateway")
		s = grpc.NewServer(grpc.Creds(creds))
	}

	pb.RegisterCryptoServiceServer(s, &cryptoServer{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("could not start server: %v\n", err)
	}
}

func (*cryptoServer) GetCryptoData(ctx context.Context, req *pb.CryptoRequest) (*pb.CryptoResponse, error) {
	redisData, err := store.CheckStore(ctx, "crypto")

	if err != redis.Nil {
		redisResp, err := store.GetCryptoData([]byte(redisData))

		if err == nil {
			return redisResp, nil
		}
	}

	url := os.Getenv("CRYPTO_THREE_URL")
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not gather data from crypto compare: %v", err)
	}

	request.Header.Add("Authorization", "Bearer "+os.Getenv("CRYPTO_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("could not make request to crypto compare: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("could not read response body: %v", err)
	}

	cryptoResp, err := handlers.HandleCryptoData(body)

	if err != nil {
		return nil, fmt.Errorf("error handling external data: %v", err)
	}

	err = store.SetStoreData(ctx, "crypto", cryptoResp)

	if err != nil {
		fmt.Printf("could not set data in redis store: %v\n", err)
	}

	return cryptoResp, nil
}

func (*cryptoServer) GetMonthlyData(ctx context.Context, req *pb.MonthlyDataRequest) (*pb.MonthlyDataResponse, error) {
	var url string

	switch {
	case req.GetCurrency() == "btc":
		url = os.Getenv("CRYPTO_BTC_MONTHLY")
	case req.GetCurrency() == "eth":
		url = os.Getenv("CRYPTO_ETH_MONTHLY")
	case req.GetCurrency() == "doge":
		url = os.Getenv("CRYPTO_DOGE_MONTHLY")
	default:
		return nil, fmt.Errorf("currency not available")
	}

	redisData, err := store.CheckStore(ctx, url+"monthly")

	if err != redis.Nil {
		redisResp, err := store.GetMonthlyData([]byte(redisData))

		if err == nil {
			return redisResp, nil
		}
	}

	if err != redis.Nil {
		redisResp, err := store.GetMonthlyData([]byte(redisData))

		if err == nil {
			return redisResp, nil
		}
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not gather data from crypto compare: %v", err)
	}

	request.Header.Add("Authorization", "Bearer "+os.Getenv("CRYPTO_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("could not make request to crypto compare: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("could not read response body: %v", err)
	}

	monthlyResp, err := handlers.HandleMonthlyData(body)

	if err != nil {
		return nil, fmt.Errorf("error handling external data: %v", err)
	}

	err = store.SetStoreData(ctx, url+"monthly", monthlyResp)

	if err != nil {
		fmt.Printf("could not set data in redis store: %v\n", err)
	}

	return monthlyResp, nil
}
