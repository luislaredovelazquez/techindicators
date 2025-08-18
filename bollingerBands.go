package techindicators

import (
	"errors"
	"fmt"
	"math"
)

// BollingerBands represents Bollinger Bands values
type BollingerBands struct {
	Timestamp  string  `json:"timestamp"`
	UpperBand  float64 `json:"upper_band"`
	MiddleBand float64 `json:"middle_band"` // This is the SMA
	LowerBand  float64 `json:"lower_band"`
	BandWidth  float64 `json:"band_width"` // (Upper - Lower) / Middle
}

// CalculateBollingerBands calculates Bollinger Bands for the given dataset
func CalculateBollingerBands(dataset [][]string, period int, multiplier float64, priceType PriceType) ([]BollingerBands, error) {
	if len(dataset) == 0 {
		return nil, errors.New("dataset is empty")
	}

	if period <= 0 {
		return nil, errors.New("period must be greater than 0")
	}

	if period > len(dataset) {
		return nil, fmt.Errorf("period (%d) cannot be greater than dataset length (%d)", period, len(dataset))
	}

	if multiplier <= 0 {
		return nil, errors.New("multiplier must be greater than 0")
	}

	var results []BollingerBands

	// Calculate Bollinger Bands for each possible position
	for i := period - 1; i < len(dataset); i++ {
		var prices []float64
		sum := 0.0

		// Collect prices for the period
		for j := i - period + 1; j <= i; j++ {
			price, err := extractPrice(dataset[j], priceType)
			if err != nil {
				return nil, fmt.Errorf("error at index %d: %w", j, err)
			}
			prices = append(prices, price)
			sum += price
		}

		// Calculate SMA (middle band)
		sma := sum / float64(period)

		// Calculate standard deviation
		varianceSum := 0.0
		for _, price := range prices {
			diff := price - sma
			varianceSum += diff * diff
		}
		stdDev := math.Sqrt(varianceSum / float64(period))

		// Calculate bands
		upperBand := sma + (multiplier * stdDev)
		lowerBand := sma - (multiplier * stdDev)

		// Calculate band width (volatility measure)
		bandWidth := 0.0
		if sma != 0 {
			bandWidth = (upperBand - lowerBand) / sma
		}

		results = append(results, BollingerBands{
			Timestamp:  dataset[i][0],
			UpperBand:  upperBand,
			MiddleBand: sma,
			LowerBand:  lowerBand,
			BandWidth:  bandWidth,
		})
	}

	return results, nil
}

// GetLatestBollingerBands returns the most recent Bollinger Bands values
func GetLatestBollingerBands(dataset [][]string, period int, multiplier float64, priceType PriceType) (BollingerBands, error) {
	bands, err := CalculateBollingerBands(dataset, period, multiplier, priceType)
	if err != nil {
		return BollingerBands{}, err
	}

	if len(bands) == 0 {
		return BollingerBands{}, errors.New("no Bollinger Bands calculated")
	}

	return bands[len(bands)-1], nil
}

// BollingerPosition indicates where the price is relative to Bollinger Bands
type BollingerPosition string

const (
	AboveUpperBand BollingerPosition = "above_upper"    // Potential overbought
	BetweenBands   BollingerPosition = "between_bands"  // Normal range
	BelowLowerBand BollingerPosition = "below_lower"    // Potential oversold
	TouchingUpper  BollingerPosition = "touching_upper" // Near upper band
	TouchingLower  BollingerPosition = "touching_lower" // Near lower band
)

// GetPricePosition determines where current price is relative to Bollinger Bands
func GetPricePosition(dataset [][]string, period int, multiplier float64, priceType PriceType, tolerance float64) (BollingerPosition, error) {
	if len(dataset) == 0 {
		return "", errors.New("dataset is empty")
	}

	// Get latest Bollinger Bands
	bands, err := GetLatestBollingerBands(dataset, period, multiplier, priceType)
	if err != nil {
		return "", err
	}

	// Get current price
	currentPrice, err := extractPrice(dataset[len(dataset)-1], ClosePrice)
	if err != nil {
		return "", err
	}

	// Calculate tolerance ranges
	upperTolerance := bands.UpperBand * (1 - tolerance)
	lowerTolerance := bands.LowerBand * (1 + tolerance)

	// Determine position
	if currentPrice > bands.UpperBand {
		return AboveUpperBand, nil
	} else if currentPrice < bands.LowerBand {
		return BelowLowerBand, nil
	} else if currentPrice >= upperTolerance {
		return TouchingUpper, nil
	} else if currentPrice <= lowerTolerance {
		return TouchingLower, nil
	}

	return BetweenBands, nil
}

// BollingerSqueeze detects if bands are in a squeeze (low volatility)
func BollingerSqueeze(dataset [][]string, period int, multiplier float64, priceType PriceType, lookback int) (bool, error) {
	bands, err := CalculateBollingerBands(dataset, period, multiplier, priceType)
	if err != nil {
		return false, err
	}

	if len(bands) < lookback {
		return false, errors.New("insufficient data for squeeze analysis")
	}

	// Get recent band widths
	recentBands := bands[len(bands)-lookback:]

	// Calculate average band width over lookback period
	totalWidth := 0.0
	for _, band := range recentBands {
		totalWidth += band.BandWidth
	}
	avgWidth := totalWidth / float64(lookback)

	// Current band width
	currentWidth := bands[len(bands)-1].BandWidth

	// Squeeze detected if current width is significantly below average
	return currentWidth < avgWidth*0.7, nil // 30% below average indicates squeeze
}

// BollingerBreakout detects potential breakouts from Bollinger Bands
func BollingerBreakout(dataset [][]string, period int, multiplier float64, priceType PriceType) (string, error) {
	if len(dataset) < 2 {
		return "insufficient_data", nil
	}

	// Get current and previous positions
	currentPos, err := GetPricePosition(dataset, period, multiplier, priceType, 0.02) // 2% tolerance
	if err != nil {
		return "", err
	}

	// Check previous candle position
	prevDataset := dataset[:len(dataset)-1]
	if len(prevDataset) < period {
		return "insufficient_data", nil
	}

	prevPos, err := GetPricePosition(prevDataset, period, multiplier, priceType, 0.02)
	if err != nil {
		return "", err
	}

	// Detect breakouts
	if prevPos == BetweenBands && currentPos == AboveUpperBand {
		return "bullish_breakout", nil
	} else if prevPos == BetweenBands && currentPos == BelowLowerBand {
		return "bearish_breakout", nil
	} else if prevPos == TouchingUpper && currentPos == AboveUpperBand {
		return "bullish_breakout", nil
	} else if prevPos == TouchingLower && currentPos == BelowLowerBand {
		return "bearish_breakout", nil
	}

	return "no_breakout", nil
}

// BollingerStrategy provides comprehensive Bollinger Bands analysis
type BollingerStrategy struct {
	Position  BollingerPosition `json:"position"`
	Breakout  string            `json:"breakout"`
	Squeeze   bool              `json:"squeeze"`
	BandWidth float64           `json:"band_width"`
	Signal    string            `json:"signal"`
}

// AnalyzeBollingerStrategy provides complete Bollinger Bands analysis for trading decisions
func AnalyzeBollingerStrategy(dataset [][]string, period int, multiplier float64, priceType PriceType) (BollingerStrategy, error) {
	position, err := GetPricePosition(dataset, period, multiplier, priceType, 0.02)
	if err != nil {
		return BollingerStrategy{}, err
	}

	breakout, err := BollingerBreakout(dataset, period, multiplier, priceType)
	if err != nil {
		return BollingerStrategy{}, err
	}

	squeeze, err := BollingerSqueeze(dataset, period, multiplier, priceType, 10)
	if err != nil {
		return BollingerStrategy{}, err
	}

	bands, err := GetLatestBollingerBands(dataset, period, multiplier, priceType)
	if err != nil {
		return BollingerStrategy{}, err
	}

	// Generate trading signal
	signal := "hold"
	switch {
	case breakout == "bullish_breakout" && !squeeze:
		signal = "strong_buy"
	case breakout == "bearish_breakout" && !squeeze:
		signal = "strong_sell"
	case position == BelowLowerBand && squeeze:
		signal = "buy" // Oversold in low volatility
	case position == AboveUpperBand && squeeze:
		signal = "sell" // Overbought in low volatility
	case squeeze && position == BetweenBands:
		signal = "wait_for_breakout"
	case position == TouchingLower:
		signal = "buy_signal"
	case position == TouchingUpper:
		signal = "sell_signal"
	}

	return BollingerStrategy{
		Position:  position,
		Breakout:  breakout,
		Squeeze:   squeeze,
		BandWidth: bands.BandWidth,
		Signal:    signal,
	}, nil
}
