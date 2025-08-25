package techindicators

import (
	"fmt"
)

// UltimateMemecoinAnalysis combines all indicators with volume confirmation
type UltimateMemecoinAnalysis struct {
	Technical     CombinedTechnicalAnalysis `json:"technical"`
	Volume        VolumeStrategy            `json:"volume"`
	FinalSignal   string                    `json:"final_signal"`
	Confidence    string                    `json:"confidence"`
	RiskLevel     string                    `json:"risk_level"`
	RugPullRisk   string                    `json:"rug_pull_risk"`  // low, medium, high, extreme
	VolumeConfirm bool                      `json:"volume_confirm"` // true if volume confirms signal
}

// UltimateAnalysis provides the most comprehensive memecoin analysis
func UltimateAnalysis(dataset []OHLCV, smaPeriod, bbPeriod, rsiPeriod, vmaPeriod int, bbMultiplier float64) (UltimateMemecoinAnalysis, error) {
	// Get technical analysis
	technical, err := ComprehensiveAnalysis(dataset, smaPeriod, bbPeriod, rsiPeriod, bbMultiplier, ClosePrice)
	if err != nil {
		return UltimateMemecoinAnalysis{}, err
	}

	// Get volume analysis
	volume, err := AnalyzeVolumeStrategy(dataset, vmaPeriod, 5)
	if err != nil {
		return UltimateMemecoinAnalysis{}, err
	}

	// Check volume confirmation
	volumeConfirm := false
	switch {
	case (technical.FinalSignal == "STRONG BUY" || technical.FinalSignal == "BUY") &&
		(volume.Signal == "strong_buy" || volume.Signal == "buy" || volume.Signal == "accumulate"):
		volumeConfirm = true
	case (technical.FinalSignal == "STRONG SELL" || technical.FinalSignal == "SELL") &&
		(volume.Signal == "strong_sell" || volume.Signal == "sell" || volume.Signal == "distribute"):
		volumeConfirm = true
	case technical.FinalSignal == "WAIT" && volume.VolumeRatio < 1.0:
		volumeConfirm = true
	}

	// Assess rug pull risk
	rugPullRisk := "low"
	switch {
	case volume.Signal == "strong_sell" && volume.AccumulationSignal.Type == "distribution" &&
		technical.RSISignal == "strong_sell" && volume.VolumeRatio > 3.0:
		rugPullRisk = "extreme"
	case volume.AccumulationSignal.Type == "distribution" && technical.FinalSignal == "STRONG SELL":
		rugPullRisk = "high"
	case volume.Signal == "distribute" || (volume.VolumeRatio > 2.0 && technical.FinalSignal == "SELL"):
		rugPullRisk = "medium"
	}

	// Adjust final signal based on volume confirmation
	finalSignal := technical.FinalSignal
	confidence := technical.Confidence
	riskLevel := technical.RiskLevel

	if volumeConfirm {
		// Volume confirms technical signal - increase confidence
		if confidence == "MEDIUM" {
			confidence = "HIGH"
		} else if confidence == "LOW" {
			confidence = "MEDIUM"
		}
	} else {
		// Volume doesn't confirm - decrease confidence and adjust signal
		if confidence == "HIGH" {
			confidence = "MEDIUM"
			if finalSignal == "STRONG BUY" {
				finalSignal = "BUY"
			} else if finalSignal == "STRONG SELL" {
				finalSignal = "SELL"
			}
		} else if confidence == "MEDIUM" {
			confidence = "LOW"
			finalSignal = "HOLD"
		}
	}

	// Special cases for volume signals
	if volume.Signal == "low_volume_alert" {
		finalSignal = "SUSPICIOUS"
		confidence = "LOW"
		riskLevel = "HIGH"
	}

	return UltimateMemecoinAnalysis{
		Technical:     technical,
		Volume:        volume,
		FinalSignal:   finalSignal,
		Confidence:    confidence,
		RiskLevel:     riskLevel,
		RugPullRisk:   rugPullRisk,
		VolumeConfirm: volumeConfirm,
	}, nil
}

// ComprehensiveAnalysis combines all indicators for ultimate trading decisions
func ComprehensiveAnalysis(dataset []OHLCV, smaPeriod, bbPeriod, rsiPeriod int, bbMultiplier float64, priceType PriceType) (CombinedTechnicalAnalysis, error) {
	// SMA Analysis
	isAboveSMA, _ := IsPriceAboveSMA(dataset, smaPeriod, priceType)
	smaCross, _ := SMACrossover(dataset, smaPeriod/2, smaPeriod, priceType)

	smaSignal := "neutral"
	if isAboveSMA && smaCross == "bullish_crossover" {
		smaSignal = "strong_bullish"
	} else if !isAboveSMA && smaCross == "bearish_crossover" {
		smaSignal = "strong_bearish"
	} else if isAboveSMA {
		smaSignal = "bullish"
	} else {
		smaSignal = "bearish"
	}

	// Bollinger Bands Analysis
	bbStrategy, _ := AnalyzeBollingerStrategy(dataset, bbPeriod, bbMultiplier, priceType)

	// RSI Analysis
	rsiStrategy, _ := AnalyzeRSIStrategy(dataset, rsiPeriod, priceType)

	// Combine signals
	signals := []string{smaSignal, bbStrategy.Signal, rsiStrategy.Signal}
	bullishCount := 0
	bearishCount := 0

	for _, signal := range signals {
		switch {
		case signal == "strong_buy" || signal == "buy" || signal == "bullish" || signal == "strong_bullish":
			bullishCount++
		case signal == "strong_sell" || signal == "sell" || signal == "bearish" || signal == "strong_bearish":
			bearishCount++
		}
	}

	// Final decision logic
	finalSignal := "HOLD"
	confidence := "LOW"
	riskLevel := "MEDIUM"

	switch {
	case bullishCount >= 3:
		finalSignal = "STRONG BUY"
		confidence = "HIGH"
		riskLevel = "LOW"
	case bullishCount >= 2:
		finalSignal = "BUY"
		confidence = "MEDIUM"
		riskLevel = "LOW"
	case bearishCount >= 3:
		finalSignal = "STRONG SELL"
		confidence = "HIGH"
		riskLevel = "HIGH"
	case bearishCount >= 2:
		finalSignal = "SELL"
		confidence = "MEDIUM"
		riskLevel = "MEDIUM"
	case bbStrategy.Signal == "wait_for_breakout":
		finalSignal = "WAIT"
		confidence = "HIGH"
		riskLevel = "LOW"
	}

	// Adjust for extreme conditions
	if rsiStrategy.Condition == RSIExtremeHigh && bbStrategy.Position == AboveUpperBand {
		finalSignal = "STRONG SELL"
		confidence = "HIGH"
		riskLevel = "HIGH"
	} else if rsiStrategy.Condition == RSIExtremeLow && bbStrategy.Position == BelowLowerBand {
		finalSignal = "STRONG BUY"
		confidence = "HIGH"
		riskLevel = "LOW"
	}

	return CombinedTechnicalAnalysis{
		SMASignal:       smaSignal,
		BollingerSignal: bbStrategy.Signal,
		RSISignal:       rsiStrategy.Signal,
		FinalSignal:     finalSignal,
		Confidence:      confidence,
		RiskLevel:       riskLevel,
	}, nil
}

// Example usage for memecoin trading
func exampleUsage() {
	// Example dataset - would need to be converted from [][]string to []OHLCV
	// This example function would need to be updated to use proper OHLCV data
	fmt.Println("Example function requires OHLCV data structure")
}
