package techindicators

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"

	"github.com/JulianToledano/goingecko/v3/api"
	"github.com/mark3labs/mcp-go/mcp"
)

type Sharpe struct {
	Coin              string  `json:"coin"`
	AvgDailyReturn    float64 `json:"avgDailyReturn"`
	DailyVolatility   float64 `json:"dailyVolatility"`
	DailySharpeRatio  float64 `json:"dailySharpeRatio"`
	AnnualSharpeRatio float64 `json:"anualSharpeRatio"`
}

// Helper: calculates average
func average(data []float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// Helper: calculates standard deviation
func stdDev(data []float64, mean float64) float64 {
	variance := 0.0
	for _, v := range data {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(len(data) - 1)
	return math.Sqrt(variance)
}

func calculateSharpeRatio(ctx context.Context, coinID, vsCurrency, days string) ([]byte, error) {
	client := api.NewDefaultClient()

	// coinID := "solana" // Replace with your chosen meme coin ID
	// vsCurrency := "usd"
	// days := "90" // Last 90 days of data

	resp, err := client.CoinsIdMarketChart(
		ctx,
		coinID,
		vsCurrency,
		days,
	)
	if err != nil {
		log.Fatalf("Error fetching market chart: %v", err)
	}

	prices := resp.Prices
	if len(prices) < 2 {
		log.Fatalf("Not enough data points for coin %s", coinID)
	}

	// Compute daily returns
	returns := make([]float64, 0, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		prev := prices[i-1][1]
		curr := prices[i][1]
		returns = append(returns, (curr-prev)/prev)
	}

	// Average return
	mean := average(returns)

	// Standard deviation
	sd := stdDev(returns, mean)

	// Risk-free rate — assuming 0 for crypto
	rf := 0.0

	dailySharpe := (mean - rf) / sd

	// Annualize assuming 365 trading days
	annualSharpe := dailySharpe * math.Sqrt(365)

	fmt.Printf("Meme Coin: %s\n", coinID)
	fmt.Printf("Avg Daily Return: %.5f\n", mean)
	fmt.Printf("Daily Volatility: %.5f\n", sd)
	fmt.Printf("Daily Sharpe Ratio: %.5f\n", dailySharpe)
	fmt.Printf("Annualized Sharpe Ratio: %.5f\n", annualSharpe)

	sharpeobj := Sharpe{
		Coin:              coinID,
		AvgDailyReturn:    mean,
		DailyVolatility:   sd,
		DailySharpeRatio:  dailySharpe,
		AnnualSharpeRatio: annualSharpe,
	}

	jsonSharpe, err := json.Marshal(sharpeobj)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return jsonSharpe, nil

}

// func calculateSharpeRatio(ctx context.Context, coinID, vsCurrency, days string) ([]byte, error) {

func SharpeRatioHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {

	coinID, err := request.RequireString("coinID")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	vsCurrency, err := request.RequireString("vsCurrency")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	days, err := request.RequireString("days")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	sharpeRatio, err := calculateSharpeRatio(ctx, coinID, vsCurrency, days)

	if err != nil {
		log.Print("error calculating sharpe ratio")
	}

	return mcp.NewToolResultText(string(sharpeRatio)), nil
}
