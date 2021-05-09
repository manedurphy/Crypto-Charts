package main

import (
	"context"
	"encoding/json"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	rdb *redis.Client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("redis-url"),
		Password: getPassword(),
		DB:       0,
	})
)

type server struct {
	pb.UnimplementedBitcoinServiceServer
}

type externalData struct {
	Bpi map[string]float64
}

func (*server) GetBitCoinData(ctx context.Context, req *pb.BitcoinRequest) (*pb.BitcoinResponse, error) {
	redisData, err := rdb.Get(ctx, "data").Result()

	if err != redis.Nil {
		redisResp, err := handleRedisData([]byte(redisData))

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

func handleRedisData(redisData []byte) (*pb.BitcoinResponse, error) {
	var data []*pb.BitcoinDatum
	err := json.Unmarshal(redisData, &data)

	if err != nil {
		return nil, fmt.Errorf("error unmarshaling data from redis cache: %v", err)
	}

	resp := []*pb.BitcoinDatum{}

	for _, v := range data {
		resp = append(resp, &pb.BitcoinDatum{Date: v.Date, Value: v.Value})
	}

	sort.Slice(resp, func(i int, j int) bool {
		return resp[i].Date < resp[j].Date
	})

	fmt.Println("Sending data from redis store!")

	return &pb.BitcoinResponse{Data: data}, nil
}

func getPassword() string {
	file, err := os.Open("/mnt/secrets-store/redis")

	if err != nil {
		return ""
	}

	defer file.Close()

	fmt.Println("successfully received secret from file system")

	secret, _ := ioutil.ReadAll(file)
	return string(secret)
}

func main() {
	lis, err := net.Listen("tcp", ":8080")

	if err != nil {
		log.Fatalf("could not listen on port 8080: %v\n", err)
	}

	fmt.Println("gRPC server started on on port 8080")

	s := grpc.NewServer()
	pb.RegisterBitcoinServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("could not start server: %v\n", err)
	}
}
