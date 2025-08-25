package techindicators

import (
	"errors"
	"fmt"
)

// CombinedTechnicalAnalysis integrates SMA, Bollinger Bands, and RSI
type CombinedTechnicalAnalysis struct {
	SMASignal       string `json:"sma_signal"`
	BollingerSignal string `json:"bollinger_signal"`
	RSISignal       string `json:"rsi_signal"`
	FinalSignal     string `json:"final_signal"`
	Confidence      string `json:"confidence"`
	RiskLevel       string `json:"risk_level"`
}

// VolumeResult represents volume analysis result
type VolumeResult struct {
	Timestamp string  `json:"timestamp"`
	Volume    float64 `json:"volume"`
	VMA       float64 `json:"vma"`  // Volume Moving Average
	OBV       float64 `json:"obv"`  // On-Balance Volume
	VPT       float64 `json:"vpt"`  // Volume Price Trend
	VROC      float64 `json:"vroc"` // Volume Rate of Change
	ADL       float64 `json:"adl"`  // Accumulation/Distribution Line
}

// VolumeSignal represents volume-based trading signals
type VolumeSignal struct {
	Type       string  `json:"type"`       // breakout, accumulation, distribution, normal
	Strength   string  `json:"strength"`   // weak, moderate, strong, extreme
	Trend      string  `json:"trend"`      // bullish, bearish, neutral
	Confidence float64 `json:"confidence"` // 0-1 scale
}

// CalculateVolumeAnalysis performs comprehensive volume analysis
func CalculateVolumeAnalysis(dataset []OHLCV, vmaPeriod, vrocPeriod int) ([]VolumeResult, error) {
	if len(dataset) == 0 {
		return nil, errors.New("dataset is empty")
	}

	if vmaPeriod <= 0 || vrocPeriod <= 0 {
		return nil, errors.New("periods must be greater than 0")
	}

	maxPeriod := vmaPeriod
	if vrocPeriod > maxPeriod {
		maxPeriod = vrocPeriod
	}

	if len(dataset) <= maxPeriod {
		return nil, fmt.Errorf("insufficient data: need more than %d candles", maxPeriod)
	}

	var results []VolumeResult
	var obv, vpt, adl float64 // Running totals

	// Extract initial data
	volumes := make([]float64, len(dataset))
	closes := make([]float64, len(dataset))
	highs := make([]float64, len(dataset))
	lows := make([]float64, len(dataset))

	for i, candle := range dataset {
		volumes[i] = candle.Volume
		closes[i] = candle.Close
		highs[i] = candle.High
		lows[i] = candle.Low
	}

	// Initialize first OBV value
	obv = volumes[0]
	vpt = volumes[0]
	adl = volumes[0]

	// Calculate indicators for each period
	for i := maxPeriod; i < len(dataset); i++ {
		// Volume Moving Average (VMA)
		vmaSum := 0.0
		for j := i - vmaPeriod + 1; j <= i; j++ {
			vmaSum += volumes[j]
		}
		vma := vmaSum / float64(vmaPeriod)

		// On-Balance Volume (OBV)
		if i > 0 {
			if closes[i] > closes[i-1] {
				obv += volumes[i]
			} else if closes[i] < closes[i-1] {
				obv -= volumes[i]
			}
			// If close unchanged, OBV unchanged
		}

		// Volume Price Trend (VPT)
		if i > 0 && closes[i-1] != 0 {
			priceChange := (closes[i] - closes[i-1]) / closes[i-1]
			vpt += volumes[i] * priceChange
		}

		// Volume Rate of Change (VROC)
		vroc := 0.0
		if i >= vrocPeriod && volumes[i-vrocPeriod] != 0 {
			vroc = ((volumes[i] - volumes[i-vrocPeriod]) / volumes[i-vrocPeriod]) * 100
		}

		// Accumulation/Distribution Line (ADL)
		if highs[i] != lows[i] {
			moneyFlowMultiplier := ((closes[i] - lows[i]) - (highs[i] - closes[i])) / (highs[i] - lows[i])
			moneyFlowVolume := moneyFlowMultiplier * volumes[i]
			adl += moneyFlowVolume
		}

		results = append(results, VolumeResult{
			Timestamp: dataset[i].Timestamp.Format("2006-01-02T15:04:05Z"),
			Volume:    volumes[i],
			VMA:       vma,
			OBV:       obv,
			VPT:       vpt,
			VROC:      vroc,
			ADL:       adl,
		})
	}

	return results, nil
}

// GetLatestVolumeAnalysis returns the most recent volume analysis
func GetLatestVolumeAnalysis(dataset []OHLCV, vmaPeriod, vrocPeriod int) (VolumeResult, error) {
	results, err := CalculateVolumeAnalysis(dataset, vmaPeriod, vrocPeriod)
	if err != nil {
		return VolumeResult{}, err
	}

	if len(results) == 0 {
		return VolumeResult{}, errors.New("no volume results calculated")
	}

	return results[len(results)-1], nil
}

// DetectVolumeBreakout identifies unusual volume activity
func DetectVolumeBreakout(dataset []OHLCV, vmaPeriod int, multiplier float64) (VolumeSignal, error) {
	latest, err := GetLatestVolumeAnalysis(dataset, vmaPeriod, 5)
	if err != nil {
		return VolumeSignal{}, err
	}

	// Compare current volume with moving average
	volumeRatio := latest.Volume / latest.VMA

	var signal VolumeSignal

	// Determine breakout strength
	switch {
	case volumeRatio >= multiplier*3:
		signal.Strength = "extreme"
		signal.Confidence = 0.9
	case volumeRatio >= multiplier*2:
		signal.Strength = "strong"
		signal.Confidence = 0.8
	case volumeRatio >= multiplier:
		signal.Strength = "moderate"
		signal.Confidence = 0.6
	default:
		signal.Strength = "weak"
		signal.Confidence = 0.3
	}

	// Determine signal type
	if volumeRatio >= multiplier {
		signal.Type = "breakout"

		// Determine trend based on price action and OBV
		if len(dataset) >= 2 {
			currentClose := dataset[len(dataset)-1].Close
			prevClose := dataset[len(dataset)-2].Close

			if currentClose > prevClose && latest.OBV > 0 {
				signal.Trend = "bullish"
			} else if currentClose < prevClose {
				signal.Trend = "bearish"
			} else {
				signal.Trend = "neutral"
			}
		}
	} else {
		signal.Type = "normal"
		signal.Trend = "neutral"
	}

	return signal, nil
}

// DetectAccumulationDistribution analyzes money flow patterns
func DetectAccumulationDistribution(dataset []OHLCV, lookback int) (VolumeSignal, error) {
	if lookback < 5 {
		lookback = 5
	}

	results, err := CalculateVolumeAnalysis(dataset, 10, 5)
	if err != nil {
		return VolumeSignal{}, err
	}

	if len(results) < lookback {
		return VolumeSignal{Type: "insufficient_data"}, nil
	}

	// Analyze recent ADL trend
	recent := results[len(results)-lookback:]

	// Calculate ADL slope (simple linear regression)
	adlSum := 0.0
	timeSum := 0.0
	adlTimeSum := 0.0
	timeSquareSum := 0.0

	for i, result := range recent {
		time := float64(i)
		adlSum += result.ADL
		timeSum += time
		adlTimeSum += result.ADL * time
		timeSquareSum += time * time
	}

	n := float64(len(recent))
	slope := (n*adlTimeSum - timeSum*adlSum) / (n*timeSquareSum - timeSum*timeSum)

	var signal VolumeSignal

	// Determine accumulation/distribution based on slope
	switch {
	case slope > 1000:
		signal.Type = "accumulation"
		signal.Strength = "strong"
		signal.Trend = "bullish"
		signal.Confidence = 0.8
	case slope > 100:
		signal.Type = "accumulation"
		signal.Strength = "moderate"
		signal.Trend = "bullish"
		signal.Confidence = 0.6
	case slope < -1000:
		signal.Type = "distribution"
		signal.Strength = "strong"
		signal.Trend = "bearish"
		signal.Confidence = 0.8
	case slope < -100:
		signal.Type = "distribution"
		signal.Strength = "moderate"
		signal.Trend = "bearish"
		signal.Confidence = 0.6
	default:
		signal.Type = "neutral"
		signal.Strength = "weak"
		signal.Trend = "neutral"
		signal.Confidence = 0.3
	}

	return signal, nil
}

// VolumeStrategy provides comprehensive volume analysis
type VolumeStrategy struct {
	Current            VolumeResult `json:"current"`
	BreakoutSignal     VolumeSignal `json:"breakout_signal"`
	AccumulationSignal VolumeSignal `json:"accumulation_signal"`
	VolumeRatio        float64      `json:"volume_ratio"` // Current volume / VMA
	OBVTrend           string       `json:"obv_trend"`    // rising, falling, sideways
	Signal             string       `json:"signal"`       // buy, sell, hold, alert
}

// AnalyzeVolumeStrategy provides complete volume analysis for trading decisions
func AnalyzeVolumeStrategy(dataset []OHLCV, vmaPeriod, vrocPeriod int) (VolumeStrategy, error) {
	// Get current volume analysis
	current, err := GetLatestVolumeAnalysis(dataset, vmaPeriod, vrocPeriod)
	if err != nil {
		return VolumeStrategy{}, err
	}

	// Detect volume breakout
	breakoutSignal, err := DetectVolumeBreakout(dataset, vmaPeriod, 2.0)
	if err != nil {
		return VolumeStrategy{}, err
	}

	// Detect accumulation/distribution
	accumSignal, err := DetectAccumulationDistribution(dataset, 10)
	if err != nil {
		return VolumeStrategy{}, err
	}

	// Calculate volume ratio
	volumeRatio := current.Volume / current.VMA

	// Determine OBV trend
	results, _ := CalculateVolumeAnalysis(dataset, vmaPeriod, vrocPeriod)
	obvTrend := "sideways"
	if len(results) >= 3 {
		recent := results[len(results)-3:]
		if recent[2].OBV > recent[1].OBV && recent[1].OBV > recent[0].OBV {
			obvTrend = "rising"
		} else if recent[2].OBV < recent[1].OBV && recent[1].OBV < recent[0].OBV {
			obvTrend = "falling"
		}
	}

	// Generate trading signal
	signal := "hold"
	switch {
	case breakoutSignal.Type == "breakout" && breakoutSignal.Trend == "bullish" && accumSignal.Type == "accumulation":
		signal = "strong_buy"
	case breakoutSignal.Type == "breakout" && breakoutSignal.Trend == "bearish" && accumSignal.Type == "distribution":
		signal = "strong_sell"
	case breakoutSignal.Type == "breakout" && breakoutSignal.Trend == "bullish":
		signal = "buy"
	case breakoutSignal.Type == "breakout" && breakoutSignal.Trend == "bearish":
		signal = "sell"
	case accumSignal.Type == "accumulation" && obvTrend == "rising":
		signal = "accumulate"
	case accumSignal.Type == "distribution" && obvTrend == "falling":
		signal = "distribute"
	case volumeRatio < 0.5:
		signal = "low_volume_alert" // Potentially fake moves
	}

	return VolumeStrategy{
		Current:            current,
		BreakoutSignal:     breakoutSignal,
		AccumulationSignal: accumSignal,
		VolumeRatio:        volumeRatio,
		OBVTrend:           obvTrend,
		Signal:             signal,
	}, nil
}
