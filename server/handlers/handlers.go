package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/manedurphy/grpc-web/pb"
)

func HandleCryptoData(body []byte) (*pb.CryptoResponse, error) {
	var externalCryptoResponse pb.ExternalCryptoResponse
	err := json.Unmarshal(body, &externalCryptoResponse)

	cryptoData := []*pb.CryptoDatum{}

	btcDatum := pb.CryptoDatum{Name: "BTC", USD: externalCryptoResponse.GetBTC().GetUSD(), EUR: externalCryptoResponse.GetBTC().GetEUR()}
	ethDatum := pb.CryptoDatum{Name: "ETH", USD: externalCryptoResponse.GetETH().GetUSD(), EUR: externalCryptoResponse.GetETH().GetEUR()}
	dogeDatum := pb.CryptoDatum{Name: "DOGE", USD: externalCryptoResponse.GetDOGE().GetUSD(), EUR: externalCryptoResponse.GetDOGE().GetEUR()}

	cryptoData = append(cryptoData, &btcDatum)
	cryptoData = append(cryptoData, &ethDatum)
	cryptoData = append(cryptoData, &dogeDatum)

	if err != nil {
		return nil, fmt.Errorf("could not handle crypto data: %v", err)
	}

	return &pb.CryptoResponse{
		Data: cryptoData,
	}, nil
}

func HandleMonthlyData(body []byte) (*pb.MonthlyDataResponse, error) {
	var monthlyData pb.MonthlyDataResponse
	err := json.Unmarshal(body, &monthlyData)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal monthly data: %v", err)
	}

	return &monthlyData, nil
}
