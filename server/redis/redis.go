package store

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/manedurphy/grpc-web/pb"
)

var (
	rdb *redis.Client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: GetPassword(),
		DB:       0,
	})
)

func GetPassword() string {
	file, err := os.Open(os.Getenv("REDIS_MOUNT_PATH"))
	// file, err := os.Open("/mnt/secrets-store/redis")

	if err != nil {
		return ""
	}

	defer file.Close()

	fmt.Println("successfully received secret from file system")

	secret, _ := ioutil.ReadAll(file)
	return string(secret)
}

func GetCryptoData(redisData []byte) (*pb.CryptoResponse, error) {
	var cryptoData *pb.CryptoResponse
	err := json.Unmarshal(redisData, &cryptoData)

	if err != nil {
		return nil, fmt.Errorf("error unmarshaling data from redis cache: %v", err)
	}

	fmt.Println("Sending data from redis store!")
	return cryptoData, nil
}

func GetMonthlyData(redisData []byte) (*pb.MonthlyDataResponse, error) {
	var monthlyData *pb.MonthlyDataResponse
	err := json.Unmarshal(redisData, &monthlyData)

	if err != nil {
		return nil, fmt.Errorf("error unmarshaling data from redis cache: %v", err)
	}

	fmt.Println("Sending data from redis store!")
	return monthlyData, nil
}

func CheckStore(ctx context.Context, key string) (string, error) {
	redisData, err := rdb.Get(ctx, key).Result()

	if err != redis.Nil {
		return redisData, nil
	}
	return "", fmt.Errorf("no data found in redis store")
}

func SetStoreData(ctx context.Context, key string, resp interface{}) error {
	cacheData, _ := json.Marshal(resp)
	err := rdb.Set(ctx, key, cacheData, 5*time.Minute).Err()

	if err != nil {
		return fmt.Errorf("could not set data in redis store: %v", err)
	}

	return nil
}
