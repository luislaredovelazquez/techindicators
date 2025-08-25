package techindicators

import (
	"errors"
	"fmt"
	"math"
)

// RSIResult represents RSI calculation result
type RSIResult struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
	Signal    string  `json:"signal"` // overbought, oversold, neutral
}

// RSICondition represents RSI market conditions
type RSICondition string

const (
	RSIOverbought  RSICondition = "overbought"   // RSI > 70
	RSIOversold    RSICondition = "oversold"     // RSI < 30
	RSINeutral     RSICondition = "neutral"      // 30 <= RSI <= 70
	RSIExtremeHigh RSICondition = "extreme_high" // RSI > 80
	RSIExtremeLow  RSICondition = "extreme_low"  // RSI < 20
)

// CalculateRSI calculates Relative Strength Index for the given dataset
func CalculateRSI(dataset []OHLCV, period int, priceType PriceType) ([]RSIResult, error) {
	if len(dataset) == 0 {
		return nil, errors.New("dataset is empty")
	}

	if period <= 0 {
		return nil, errors.New("period must be greater than 0")
	}

	if period >= len(dataset) {
		return nil, fmt.Errorf("period (%d) must be less than dataset length (%d)", period, len(dataset))
	}

	// Extract prices
	var prices []float64
	for _, candle := range dataset {
		price := candle.ExtractPrice(priceType)
		prices = append(prices, price)
	}

	var results []RSIResult

	// Calculate price changes
	var gains []float64
	var losses []float64

	for i := 1; i < len(prices); i++ {
		change := prices[i] - prices[i-1]
		if change > 0 {
			gains = append(gains, change)
			losses = append(losses, 0)
		} else {
			gains = append(gains, 0)
			losses = append(losses, -change)
		}
	}

	// Need enough data for initial calculation
	if len(gains) < period {
		return nil, fmt.Errorf("insufficient data: need at least %d price changes", period)
	}

	// Calculate initial average gain and loss (SMA for first calculation)
	var avgGain, avgLoss float64
	for i := 0; i < period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate first RSI
	rs := avgGain / avgLoss
	if avgLoss == 0 {
		rs = 100 // Avoid division by zero
	}
	rsi := 100 - (100 / (1 + rs))

	// Add first RSI result
	signal := getRSISignal(rsi)
	results = append(results, RSIResult{
		Timestamp: dataset[period].Timestamp.Format("2006-01-02T15:04:05Z"), // period+1 index in original dataset
		Value:     rsi,
		Signal:    signal,
	})

	// Calculate subsequent RSI values using smoothed averages (EMA-like)
	for i := period; i < len(gains); i++ {
		// Smoothed averages (Wilder's smoothing)
		avgGain = ((avgGain * float64(period-1)) + gains[i]) / float64(period)
		avgLoss = ((avgLoss * float64(period-1)) + losses[i]) / float64(period)

		// Calculate RSI
		rs = avgGain / avgLoss
		if avgLoss == 0 {
			rs = 100
		}
		rsi = 100 - (100 / (1 + rs))

		signal = getRSISignal(rsi)
		results = append(results, RSIResult{
			Timestamp: dataset[i+1].Timestamp.Format("2006-01-02T15:04:05Z"), // i+1 because gains array is offset by 1
			Value:     rsi,
			Signal:    signal,
		})
	}

	return results, nil
}

// getRSISignal determines the signal based on RSI value
func getRSISignal(rsi float64) string {
	switch {
	case rsi >= 80:
		return "extreme_overbought"
	case rsi >= 70:
		return "overbought"
	case rsi <= 20:
		return "extreme_oversold"
	case rsi <= 30:
		return "oversold"
	default:
		return "neutral"
	}
}

// GetLatestRSI returns the most recent RSI value
func GetLatestRSI(dataset []OHLCV, period int, priceType PriceType) (RSIResult, error) {
	rsiResults, err := CalculateRSI(dataset, period, priceType)
	if err != nil {
		return RSIResult{}, err
	}

	if len(rsiResults) == 0 {
		return RSIResult{}, errors.New("no RSI results calculated")
	}

	return rsiResults[len(rsiResults)-1], nil
}

// RSIDivergence detects bullish/bearish divergences between price and RSI
type RSIDivergence struct {
	Type       string  `json:"type"`       // bullish, bearish, none
	Strength   string  `json:"strength"`   // regular, hidden
	Confidence float64 `json:"confidence"` // 0-1 scale
}

// DetectRSIDivergence identifies potential trend reversal signals
func DetectRSIDivergence(dataset []OHLCV, period int, priceType PriceType, lookback int) (RSIDivergence, error) {
	if lookback < 5 {
		lookback = 5 // Minimum lookback for meaningful divergence
	}

	rsiResults, err := CalculateRSI(dataset, period, priceType)
	if err != nil {
		return RSIDivergence{}, err
	}

	if len(rsiResults) < lookback || len(dataset) < lookback {
		return RSIDivergence{Type: "none", Strength: "insufficient_data", Confidence: 0}, nil
	}

	// Get recent data
	recentRSI := rsiResults[len(rsiResults)-lookback:]
	recentPrices := dataset[len(dataset)-lookback:]

	// Find price and RSI extremes
	var priceHighs, priceLows []float64
	var rsiHighs, rsiLows []float64

	for i, rsi := range recentRSI {
		price := recentPrices[i].ExtractPrice(ClosePrice)

		// Simple peak/trough detection
		if i > 0 && i < len(recentRSI)-1 {
			prevRSI := recentRSI[i-1].Value
			nextRSI := recentRSI[i+1].Value

			// RSI peaks
			if rsi.Value > prevRSI && rsi.Value > nextRSI {
				rsiHighs = append(rsiHighs, rsi.Value)
				priceHighs = append(priceHighs, price)
			}

			// RSI troughs
			if rsi.Value < prevRSI && rsi.Value < nextRSI {
				rsiLows = append(rsiLows, rsi.Value)
				priceLows = append(priceLows, price)
			}
		}
	}

	// Analyze divergences
	if len(priceHighs) >= 2 && len(rsiHighs) >= 2 {
		// Bearish divergence: price makes higher highs, RSI makes lower highs
		lastPriceHigh := priceHighs[len(priceHighs)-1]
		prevPriceHigh := priceHighs[len(priceHighs)-2]
		lastRSIHigh := rsiHighs[len(rsiHighs)-1]
		prevRSIHigh := rsiHighs[len(rsiHighs)-2]

		if lastPriceHigh > prevPriceHigh && lastRSIHigh < prevRSIHigh {
			confidence := math.Abs(lastRSIHigh-prevRSIHigh) / 10.0 // Simple confidence calculation
			if confidence > 1.0 {
				confidence = 1.0
			}
			return RSIDivergence{
				Type:       "bearish",
				Strength:   "regular",
				Confidence: confidence,
			}, nil
		}
	}

	if len(priceLows) >= 2 && len(rsiLows) >= 2 {
		// Bullish divergence: price makes lower lows, RSI makes higher lows
		lastPriceLow := priceLows[len(priceLows)-1]
		prevPriceLow := priceLows[len(priceLows)-2]
		lastRSILow := rsiLows[len(rsiLows)-1]
		prevRSILow := rsiLows[len(rsiLows)-2]

		if lastPriceLow < prevPriceLow && lastRSILow > prevRSILow {
			confidence := math.Abs(lastRSILow-prevRSILow) / 10.0
			if confidence > 1.0 {
				confidence = 1.0
			}
			return RSIDivergence{
				Type:       "bullish",
				Strength:   "regular",
				Confidence: confidence,
			}, nil
		}
	}

	return RSIDivergence{Type: "none", Strength: "none", Confidence: 0}, nil
}

// RSIStrategy provides comprehensive RSI analysis
type RSIStrategy struct {
	Current    RSIResult     `json:"current"`
	Condition  RSICondition  `json:"condition"`
	Divergence RSIDivergence `json:"divergence"`
	Signal     string        `json:"signal"`
	Momentum   string        `json:"momentum"` // strengthening, weakening, neutral
}

// AnalyzeRSIStrategy provides complete RSI analysis for trading decisions
func AnalyzeRSIStrategy(dataset []OHLCV, period int, priceType PriceType) (RSIStrategy, error) {
	// Get current RSI
	currentRSI, err := GetLatestRSI(dataset, period, priceType)
	if err != nil {
		return RSIStrategy{}, err
	}

	// Determine condition
	var condition RSICondition
	switch {
	case currentRSI.Value >= 80:
		condition = RSIExtremeHigh
	case currentRSI.Value >= 70:
		condition = RSIOverbought
	case currentRSI.Value <= 20:
		condition = RSIExtremeLow
	case currentRSI.Value <= 30:
		condition = RSIOversold
	default:
		condition = RSINeutral
	}

	// Detect divergence
	divergence, err := DetectRSIDivergence(dataset, period, priceType, 10)
	if err != nil {
		return RSIStrategy{}, err
	}

	// Analyze momentum trend
	rsiResults, err := CalculateRSI(dataset, period, priceType)
	if err != nil {
		return RSIStrategy{}, err
	}

	momentum := "neutral"
	if len(rsiResults) >= 3 {
		recent := rsiResults[len(rsiResults)-3:]
		if recent[2].Value > recent[1].Value && recent[1].Value > recent[0].Value {
			momentum = "strengthening"
		} else if recent[2].Value < recent[1].Value && recent[1].Value < recent[0].Value {
			momentum = "weakening"
		}
	}

	// Generate trading signal
	signal := "hold"
	switch {
	case condition == RSIExtremeLow && divergence.Type == "bullish":
		signal = "strong_buy"
	case condition == RSIExtremeHigh && divergence.Type == "bearish":
		signal = "strong_sell"
	case condition == RSIOversold && momentum == "strengthening":
		signal = "buy"
	case condition == RSIOverbought && momentum == "weakening":
		signal = "sell"
	case currentRSI.Value > 50 && momentum == "strengthening":
		signal = "bullish"
	case currentRSI.Value < 50 && momentum == "weakening":
		signal = "bearish"
	}

	return RSIStrategy{
		Current:    currentRSI,
		Condition:  condition,
		Divergence: divergence,
		Signal:     signal,
		Momentum:   momentum,
	}, nil
}
