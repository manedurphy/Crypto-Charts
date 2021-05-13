package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/manedurphy/grpc-web/pb"
)

func HandleCryptoData(body []byte) (*pb.CryptoResponse, error) {
	var cryptoData pb.CryptoResponse
	err := json.Unmarshal(body, &cryptoData)

	if err != nil {
		return nil, err
	}

	return &cryptoData, nil
}

func HandleMonthlyData(body []byte) (*pb.MonthlyDataResponse, error) {
	var monthlyData pb.MonthlyDataResponse
	err := json.Unmarshal(body, &monthlyData)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal monthly data: %v", err)
	}

	return &monthlyData, nil
}
