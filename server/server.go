package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	pb "github.com/manedurphy/grpc-web/pb"
	store "github.com/manedurphy/grpc-web/server/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

var (
	rdb *redis.Client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: store.GetPassword(),
		DB:       0,
	})
	certfile = "tls/server.crt"
	keyfile  = "tls/server.key"

	secure = flag.Bool("secure", true, "set to true to use TLS connection")
)

type cryptoServer struct {
	pb.UnimplementedCryptoServiceServer
}

type btcServer struct {
	pb.UnimplementedBitcoinServiceServer
}

type externalData struct {
	Bpi map[string]float64
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
	pb.RegisterBitcoinServiceServer(s, &btcServer{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("could not start server: %v\n", err)
	}
}

func (*btcServer) GetBitCoinData(ctx context.Context, req *pb.BitcoinRequest) (*pb.BitcoinResponse, error) {
	redisData, err := rdb.Get(ctx, "data").Result()

	if err != redis.Nil {
		redisResp, err := store.HandleRedisData([]byte(redisData))

		if err == nil {
			return redisResp, nil
		}
	}

	resp, err := http.Get("https://api.coindesk.com/v1/bpi/historical/close.json")

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "could not get data from external api: %v", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading external data: %v", err)
	}

	btcResp, err := handleExternalData(body)

	if err != nil {
		return nil, fmt.Errorf("error handling external data: %v", err)
	}

	cacheData, _ := json.Marshal(btcResp)
	err = rdb.Set(ctx, "data", cacheData, 5*time.Minute).Err()

	if err != nil {
		fmt.Printf("could not set data in redis store: %v\n", err)
	}

	return &pb.BitcoinResponse{Data: btcResp}, nil
}

func (*cryptoServer) GetCryptoData(context.Context, *pb.CryptoRequest) (*pb.CryptoResponse, error) {
	url := "https://min-api.cryptocompare.com/data/pricemulti?fsyms=BTC,ETH,DOGE&tsyms=USD,EUR"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", "Bearer "+os.Getenv("CRYPTO_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	cryptoResp, err := handleCryptoData(body)

	if err != nil {
		return nil, err
	}

	return cryptoResp, nil
}

func handleCryptoData(body []byte) (*pb.CryptoResponse, error) {
	var cryptoData pb.CryptoResponse
	err := json.Unmarshal(body, &cryptoData)

	if err != nil {
		return nil, err
	}

	return &cryptoData, nil
}

func handleExternalData(body []byte) ([]*pb.BitcoinDatum, error) {
	var data externalData
	err := json.Unmarshal(body, &data)

	btcResp := []*pb.BitcoinDatum{}

	for k, v := range data.Bpi {
		btcResp = append(btcResp, &pb.BitcoinDatum{Date: k, Value: v})
	}

	sort.Slice(btcResp, func(i int, j int) bool {
		return btcResp[i].Date < btcResp[j].Date
	})

	if err != nil {
		log.Fatalf("could unmarshal external data: %v\n", err)
	}

	return btcResp, nil
}
