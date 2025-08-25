package techindicators

import (
	"errors"
	"fmt"
)

// PriceType represents which price to use for SMA calculation
type PriceType int

const (
	ClosePrice PriceType = iota
	OpenPrice
	HighPrice
	LowPrice
	TypicalPrice  // (High + Low + Close) / 3
	WeightedPrice // (High + Low + 2*Close) / 4
)

// SMAResult represents the result of SMA calculation
type SMAResult struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

// CalculateSMA calculates Simple Moving Average for the given dataset
func CalculateSMA(dataset []OHLCV, period int, priceType PriceType) ([]SMAResult, error) {
	if len(dataset) == 0 {
		return nil, errors.New("dataset is empty")
	}

	if period <= 0 {
		return nil, errors.New("period must be greater than 0")
	}

	if period > len(dataset) {
		return nil, fmt.Errorf("period (%d) cannot be greater than dataset length (%d)", period, len(dataset))
	}

	var results []SMAResult

	// Calculate SMA for each possible position
	for i := period - 1; i < len(dataset); i++ {
		sum := 0.0

		// Sum the last 'period' values
		for j := i - period + 1; j <= i; j++ {
			price := dataset[j].ExtractPrice(priceType)
			sum += price
		}

		// Calculate average
		smaValue := sum / float64(period)

		// Add result with corresponding timestamp
		results = append(results, SMAResult{
			Timestamp: dataset[i].Timestamp.Format("2006-01-02T15:04:05Z"),
			Value:     smaValue,
		})
	}

	return results, nil
}

// CalculateMultipleSMA calculates multiple SMAs with different periods
func CalculateMultipleSMA(dataset []OHLCV, periods []int, priceType PriceType) (map[int][]SMAResult, error) {
	results := make(map[int][]SMAResult)

	for _, period := range periods {
		sma, err := CalculateSMA(dataset, period, priceType)
		if err != nil {
			return nil, fmt.Errorf("error calculating SMA-%d: %w", period, err)
		}
		results[period] = sma
	}

	return results, nil
}

// GetLatestSMA returns the most recent SMA value
func GetLatestSMA(dataset []OHLCV, period int, priceType PriceType) (float64, error) {
	smaResults, err := CalculateSMA(dataset, period, priceType)
	if err != nil {
		return 0, err
	}

	if len(smaResults) == 0 {
		return 0, errors.New("no SMA results calculated")
	}

	return smaResults[len(smaResults)-1].Value, nil
}

// IsPriceAboveSMA checks if current price is above the SMA
func IsPriceAboveSMA(dataset []OHLCV, period int, priceType PriceType) (bool, error) {
	if len(dataset) == 0 {
		return false, errors.New("dataset is empty")
	}

	// Get latest SMA
	latestSMA, err := GetLatestSMA(dataset, period, priceType)
	if err != nil {
		return false, err
	}

	// Get current price (latest close)
	currentPrice := dataset[len(dataset)-1].ExtractPrice(ClosePrice)

	return currentPrice > latestSMA, nil
}

// SMACrossover detects if there's a bullish/bearish crossover between two SMAs
func SMACrossover(dataset []OHLCV, fastPeriod, slowPeriod int, priceType PriceType) (string, error) {
	if fastPeriod >= slowPeriod {
		return "", errors.New("fast period must be less than slow period")
	}

	if len(dataset) < slowPeriod+1 {
		return "", errors.New("insufficient data for crossover analysis")
	}

	// Calculate both SMAs
	fastSMA, err := CalculateSMA(dataset, fastPeriod, priceType)
	if err != nil {
		return "", err
	}

	slowSMA, err := CalculateSMA(dataset, slowPeriod, priceType)
	if err != nil {
		return "", err
	}

	// Need at least 2 points to detect crossover
	if len(fastSMA) < 2 || len(slowSMA) < 2 {
		return "no_signal", nil
	}

	// Get current and previous values (aligned by timestamp)
	fastCurrent := fastSMA[len(fastSMA)-1].Value
	fastPrevious := fastSMA[len(fastSMA)-2].Value
	slowCurrent := slowSMA[len(slowSMA)-1].Value
	slowPrevious := slowSMA[len(slowSMA)-2].Value

	// Check for crossover
	if fastPrevious <= slowPrevious && fastCurrent > slowCurrent {
		return "bullish_crossover", nil
	} else if fastPrevious >= slowPrevious && fastCurrent < slowCurrent {
		return "bearish_crossover", nil
	}

	return "no_signal", nil
}
