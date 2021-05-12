package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/manedurphy/grpc-web/pb"
)

func GetPassword() string {
	file, err := os.Open("/mnt/secrets-store/redis")

	if err != nil {
		return ""
	}

	defer file.Close()

	fmt.Println("successfully received secret from file system")

	secret, _ := ioutil.ReadAll(file)
	return string(secret)
}

func HandleRedisData(redisData []byte) (*pb.BitcoinResponse, error) {
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
