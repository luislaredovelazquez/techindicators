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
func UltimateAnalysis(dataset [][]string, smaPeriod, bbPeriod, rsiPeriod, vmaPeriod int, bbMultiplier float64) (UltimateMemecoinAnalysis, error) {
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
func ComprehensiveAnalysis(dataset [][]string, smaPeriod, bbPeriod, rsiPeriod int, bbMultiplier float64, priceType PriceType) (CombinedTechnicalAnalysis, error) {
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
	// Example dataset
	dataset := [][]string{
		{"1692230400", "0.001", "0.0012", "0.0013", "0.0009", "1000000"},
		{"1692234000", "0.0012", "0.0011", "0.0014", "0.001", "1200000"},
		{"1692237600", "0.0011", "0.0013", "0.0015", "0.001", "900000"},
		{"1692241200", "0.0013", "0.0014", "0.0016", "0.0012", "1100000"},
		{"1692244800", "0.0014", "0.0012", "0.0015", "0.0011", "800000"},
		{"1692248400", "0.0012", "0.0015", "0.0016", "0.0011", "1300000"},
		{"1692252000", "0.0015", "0.0016", "0.0018", "0.0014", "1500000"},
		{"1692255600", "0.0016", "0.0014", "0.0017", "0.0013", "1100000"},
		{"1692259200", "0.0014", "0.0017", "0.0019", "0.0013", "1400000"},
		{"1692262800", "0.0017", "0.0018", "0.002", "0.0016", "1600000"},
	}

	fmt.Println("=== SMA Calculation Examples ===")

	// Example 1: Calculate SMA-5 using close prices
	sma5, err := CalculateSMA(dataset, 5, ClosePrice)
	if err != nil {
		fmt.Printf("Error calculating SMA-5: %v\n", err)
		return
	}

	fmt.Println("\nSMA-5 (Close Prices):")
	for _, result := range sma5 {
		fmt.Printf("Timestamp: %s, SMA-5: %.6f\n", result.Timestamp, result.Value)
	}

	// Example 2: Get latest SMA value
	latestSMA10, err := GetLatestSMA(dataset, 5, ClosePrice)
	if err != nil {
		fmt.Printf("Error getting latest SMA: %v\n", err)
		return
	}
	fmt.Printf("\nLatest SMA-5: %.6f\n", latestSMA10)

	// Example 3: Check if price is above SMA (bullish signal)
	isAbove, err := IsPriceAboveSMA(dataset, 5, ClosePrice)
	if err != nil {
		fmt.Printf("Error checking price vs SMA: %v\n", err)
		return
	}
	fmt.Printf("Is current price above SMA-5? %v\n", isAbove)

	// Example 4: Multiple SMAs for comprehensive analysis
	periods := []int{5, 10}
	if len(dataset) >= 10 {
		multipleSMA, err := CalculateMultipleSMA(dataset, periods, ClosePrice)
		if err != nil {
			fmt.Printf("Error calculating multiple SMAs: %v\n", err)
			return
		}

		fmt.Println("\n=== Multiple SMA Analysis ===")
		for period, results := range multipleSMA {
			if len(results) > 0 {
				latest := results[len(results)-1]
				fmt.Printf("Latest SMA-%d: %.6f\n", period, latest.Value)
			}
		}
	}

	// Example 5: SMA Crossover detection (great for entry/exit signals)
	if len(dataset) >= 10 {
		crossover, err := SMACrossover(dataset, 5, 8, ClosePrice)
		if err != nil {
			fmt.Printf("Error detecting crossover: %v\n", err)
			return
		}
		fmt.Printf("\nSMA Crossover Signal (5 vs 8): %s\n", crossover)
	}

	fmt.Println("\n=== Bollinger Bands Analysis ===")

	// Example 6: Calculate Bollinger Bands (20 period, 2.0 multiplier - standard settings)
	bb20, err := CalculateBollingerBands(dataset, 5, 2.0, ClosePrice)
	if err != nil {
		fmt.Printf("Error calculating Bollinger Bands: %v\n", err)
		return
	}

	fmt.Println("\nBollinger Bands (5 period, 2.0 multiplier):")
	for i, bb := range bb20 {
		if i >= 3 { // Show last 3 entries
			fmt.Printf("Timestamp: %s\n", bb.Timestamp)
			fmt.Printf("  Upper: %.6f, Middle: %.6f, Lower: %.6f\n", bb.UpperBand, bb.MiddleBand, bb.LowerBand)
			fmt.Printf("  Band Width: %.4f\n", bb.BandWidth)
		}
	}

	// Example 7: Get current price position relative to bands
	position, err := GetPricePosition(dataset, 5, 2.0, ClosePrice, 0.02)
	if err != nil {
		fmt.Printf("Error getting price position: %v\n", err)
		return
	}
	fmt.Printf("\nCurrent Price Position: %s\n", position)

	// Example 8: Check for Bollinger Squeeze (low volatility before breakout)
	squeeze, err := BollingerSqueeze(dataset, 5, 2.0, ClosePrice, 5)
	if err != nil {
		fmt.Printf("Error checking squeeze: %v\n", err)
		return
	}
	fmt.Printf("Bollinger Squeeze Detected: %v\n", squeeze)

	// Example 9: Detect breakouts
	breakout, err := BollingerBreakout(dataset, 5, 2.0, ClosePrice)
	if err != nil {
		fmt.Printf("Error detecting breakout: %v\n", err)
		return
	}
	fmt.Printf("Breakout Signal: %s\n", breakout)

	// Example 10: Complete Bollinger Strategy Analysis
	strategy, err := AnalyzeBollingerStrategy(dataset, 5, 2.0, ClosePrice)
	if err != nil {
		fmt.Printf("Error analyzing strategy: %v\n", err)
		return
	}

	fmt.Println("\n=== Complete Bollinger Strategy ===")
	fmt.Printf("Position: %s\n", strategy.Position)
	fmt.Printf("Breakout: %s\n", strategy.Breakout)
	fmt.Printf("Squeeze: %v\n", strategy.Squeeze)
	fmt.Printf("Band Width: %.4f\n", strategy.BandWidth)
	fmt.Printf("Trading Signal: %s\n", strategy.Signal)

	// Example 11: Combined SMA + Bollinger Bands strategy
	fmt.Println("\n=== Combined SMA + Bollinger Strategy ===")

	// Get SMA trend
	isAboveSMA, _ := IsPriceAboveSMA(dataset, 5, ClosePrice)
	smaCrossover, _ := SMACrossover(dataset, 3, 5, ClosePrice)

	// Combine signals for comprehensive analysis
	fmt.Printf("Price above SMA-5: %v\n", isAboveSMA)
	fmt.Printf("SMA Crossover: %s\n", smaCrossover)
	fmt.Printf("Bollinger Signal: %s\n", strategy.Signal)

	// Generate final recommendation
	finalSignal := "HOLD"
	confidence := "LOW"

	switch {
	case strategy.Signal == "strong_buy" && isAboveSMA && smaCrossover == "bullish_crossover":
		finalSignal = "STRONG BUY"
		confidence = "HIGH"
	case strategy.Signal == "strong_sell" && !isAboveSMA && smaCrossover == "bearish_crossover":
		finalSignal = "STRONG SELL"
		confidence = "HIGH"
	case strategy.Signal == "buy" && isAboveSMA:
		finalSignal = "BUY"
		confidence = "MEDIUM"
	case strategy.Signal == "sell" && !isAboveSMA:
		finalSignal = "SELL"
		confidence = "MEDIUM"
	case strategy.Signal == "wait_for_breakout":
		finalSignal = "WAIT"
		confidence = "HIGH"
	}

	fmt.Printf("\n🎯 FINAL RECOMMENDATION: %s (Confidence: %s)\n", finalSignal, confidence)

	fmt.Println("\n=== RSI Analysis ===")

	// Example 12: Calculate RSI (14 period - standard setting)
	rsi14, err := CalculateRSI(dataset, 5, ClosePrice) // Using 5 for demo with limited data
	if err != nil {
		fmt.Printf("Error calculating RSI: %v\n", err)
		return
	}

	fmt.Println("\nRSI (5 period):")
	for i, rsi := range rsi14 {
		if i >= len(rsi14)-3 { // Show last 3 entries
			fmt.Printf("Timestamp: %s, RSI: %.2f, Signal: %s\n", rsi.Timestamp, rsi.Value, rsi.Signal)
		}
	}

	// Example 13: Get current RSI
	latestRSI, err := GetLatestRSI(dataset, 5, ClosePrice)
	if err != nil {
		fmt.Printf("Error getting latest RSI: %v\n", err)
		return
	}
	fmt.Printf("\nCurrent RSI: %.2f (%s)\n", latestRSI.Value, latestRSI.Signal)

	// Example 14: RSI Divergence Analysis
	divergence, err := DetectRSIDivergence(dataset, 5, ClosePrice, 8)
	if err != nil {
		fmt.Printf("Error detecting RSI divergence: %v\n", err)
		return
	}
	fmt.Printf("RSI Divergence: %s %s (Confidence: %.2f)\n",
		divergence.Type, divergence.Strength, divergence.Confidence)

	// Example 15: Complete RSI Strategy
	rsiStrategy, err := AnalyzeRSIStrategy(dataset, 5, ClosePrice)
	if err != nil {
		fmt.Printf("Error analyzing RSI strategy: %v\n", err)
		return
	}

	fmt.Println("\n=== Complete RSI Strategy ===")
	fmt.Printf("RSI Value: %.2f\n", rsiStrategy.Current.Value)
	fmt.Printf("Condition: %s\n", rsiStrategy.Condition)
	fmt.Printf("Momentum: %s\n", rsiStrategy.Momentum)
	fmt.Printf("Divergence: %s\n", rsiStrategy.Divergence.Type)
	fmt.Printf("RSI Signal: %s\n", rsiStrategy.Signal)

	fmt.Println("\n=== 🚀 ULTIMATE MEMECOIN ANALYSIS 🚀 ===")

	// Example 16: Comprehensive Analysis (All Indicators Combined)
	comprehensive, err := ComprehensiveAnalysis(dataset, 5, 5, 5, 2.0, ClosePrice)
	if err != nil {
		fmt.Printf("Error in comprehensive analysis: %v\n", err)
		return
	}

	fmt.Printf("📊 SMA Signal: %s\n", comprehensive.SMASignal)
	fmt.Printf("📈 Bollinger Signal: %s\n", comprehensive.BollingerSignal)
	fmt.Printf("⚡ RSI Signal: %s\n", comprehensive.RSISignal)
	fmt.Printf("\n🎯 FINAL SIGNAL: %s\n", comprehensive.FinalSignal)
	fmt.Printf("🔥 Confidence: %s\n", comprehensive.Confidence)
	fmt.Printf("⚠️  Risk Level: %s\n", comprehensive.RiskLevel)

	// Example 17: Trading Decision Framework
	fmt.Println("\n=== 💰 TRADING DECISION FRAMEWORK 💰 ===")

	switch comprehensive.FinalSignal {
	case "STRONG BUY":
		fmt.Println("✅ EXECUTE BUY ORDER")
		fmt.Println("   • All indicators align bullishly")
		fmt.Println("   • High probability setup")
		fmt.Println("   • Consider larger position size")

	case "BUY":
		fmt.Println("✅ CONSIDER BUY ORDER")
		fmt.Println("   • Majority indicators bullish")
		fmt.Println("   • Normal position size")
		fmt.Println("   • Monitor closely")

	case "STRONG SELL":
		fmt.Println("❌ EXECUTE SELL ORDER")
		fmt.Println("   • All indicators align bearishly")
		fmt.Println("   • High probability of decline")
		fmt.Println("   • Consider stop-loss if holding")

	case "SELL":
		fmt.Println("❌ CONSIDER SELL ORDER")
		fmt.Println("   • Majority indicators bearish")
		fmt.Println("   • Reduce position or take profits")
		fmt.Println("   • Set tight stop-loss")

	case "WAIT":
		fmt.Println("⏳ WAIT FOR BETTER SETUP")
		fmt.Println("   • Low volatility detected")
		fmt.Println("   • Prepare for breakout")
		fmt.Println("   • Set alerts for movement")

	case "HOLD":
		fmt.Println("🤔 MAINTAIN CURRENT POSITION")
		fmt.Println("   • Mixed signals")
		fmt.Println("   • No clear direction")
		fmt.Println("   • Wait for confirmation")
	}

	// Example 18: Risk Management Suggestions
	fmt.Println("\n=== 🛡️ RISK MANAGEMENT 🛡️ ===")

	switch comprehensive.RiskLevel {
	case "LOW":
		fmt.Println("• Position Size: Up to 3-5% of portfolio")
		fmt.Println("• Stop Loss: 8-10% below entry")
		fmt.Println("• Take Profit: 15-25% above entry")

	case "MEDIUM":
		fmt.Println("• Position Size: Up to 2-3% of portfolio")
		fmt.Println("• Stop Loss: 5-8% below entry")
		fmt.Println("• Take Profit: 10-20% above entry")

	case "HIGH":
		fmt.Println("• Position Size: Up to 1-2% of portfolio")
		fmt.Println("• Stop Loss: 3-5% below entry")
		fmt.Println("• Take Profit: 8-15% above entry")
		fmt.Println("• ⚠️ EXTREME CAUTION ADVISED")
	}

	fmt.Println("\n💡 Remember: This is a demonstration with limited data.")
	fmt.Println("💡 In live trading, use at least 100+ candles for reliable signals!")
	fmt.Println("💡 Always combine with volume analysis and market sentiment!")

	fmt.Println("\n=== 📊 VOLUME ANALYSIS ===")

	// Example 19: Volume Analysis (VMA, OBV, VPT, ADL, VROC)
	volumeResults, err := CalculateVolumeAnalysis(dataset, 5, 3)
	if err != nil {
		fmt.Printf("Error calculating volume analysis: %v\n", err)
		return
	}

	fmt.Println("\nVolume Indicators (last 3 periods):")
	for i, vol := range volumeResults {
		if i >= len(volumeResults)-3 {
			fmt.Printf("Timestamp: %s\n", vol.Timestamp)
			fmt.Printf("  Volume: %.0f, VMA: %.0f, Ratio: %.2f\n", vol.Volume, vol.VMA, vol.Volume/vol.VMA)
			fmt.Printf("  OBV: %.0f, VPT: %.2f, ADL: %.2f, VROC: %.1f%%\n",
				vol.OBV, vol.VPT, vol.ADL, vol.VROC)
		}
	}

	// Example 20: Volume Breakout Detection
	volumeBreakout, err := DetectVolumeBreakout(dataset, 5, 2.0)
	if err != nil {
		fmt.Printf("Error detecting volume breakout: %v\n", err)
		return
	}

	fmt.Printf("\nVolume Breakout Analysis:\n")
	fmt.Printf("Type: %s, Strength: %s, Trend: %s\n",
		volumeBreakout.Type, volumeBreakout.Strength, volumeBreakout.Trend)
	fmt.Printf("Confidence: %.1f%%\n", volumeBreakout.Confidence*100)

	// Example 21: Accumulation/Distribution Analysis
	accumulationSignal, err := DetectAccumulationDistribution(dataset, 8)
	if err != nil {
		fmt.Printf("Error detecting accumulation/distribution: %v\n", err)
		return
	}

	fmt.Printf("\nAccumulation/Distribution Analysis:\n")
	fmt.Printf("Type: %s, Strength: %s, Trend: %s\n",
		accumulationSignal.Type, accumulationSignal.Strength, accumulationSignal.Trend)
	fmt.Printf("Confidence: %.1f%%\n", accumulationSignal.Confidence*100)

	// Example 22: Complete Volume Strategy
	volumeStrategy, err := AnalyzeVolumeStrategy(dataset, 5, 3)
	if err != nil {
		fmt.Printf("Error analyzing volume strategy: %v\n", err)
		return
	}

	fmt.Println("\n=== Complete Volume Strategy ===")
	fmt.Printf("Current Volume Ratio: %.2f (%.0f vs %.0f VMA)\n",
		volumeStrategy.VolumeRatio, volumeStrategy.Current.Volume, volumeStrategy.Current.VMA)
	fmt.Printf("OBV Trend: %s\n", volumeStrategy.OBVTrend)
	fmt.Printf("Breakout Signal: %s (%s)\n", volumeStrategy.BreakoutSignal.Type, volumeStrategy.BreakoutSignal.Strength)
	fmt.Printf("Accumulation Signal: %s (%s)\n", volumeStrategy.AccumulationSignal.Type, volumeStrategy.AccumulationSignal.Strength)
	fmt.Printf("Volume Signal: %s\n", volumeStrategy.Signal)

	fmt.Println("\n=== 🔥 ULTIMATE MEMECOIN ANALYSIS 🔥 ===")

	// Example 23: Ultimate Analysis (All Indicators + Volume Confirmation)
	ultimate, err := UltimateAnalysis(dataset, 5, 5, 5, 5, 2.0)
	if err != nil {
		fmt.Printf("Error in ultimate analysis: %v\n", err)
		return
	}

	fmt.Printf("📈 Technical Signal: %s (%s confidence)\n", ultimate.Technical.FinalSignal, ultimate.Technical.Confidence)
	fmt.Printf("📊 Volume Signal: %s\n", ultimate.Volume.Signal)
	fmt.Printf("✅ Volume Confirms Technical: %v\n", ultimate.VolumeConfirm)
	fmt.Printf("\n🚨 RUG PULL RISK: %s\n", ultimate.RugPullRisk)
	fmt.Printf("🎯 ULTIMATE SIGNAL: %s\n", ultimate.FinalSignal)
	fmt.Printf("🔥 Final Confidence: %s\n", ultimate.Confidence)
	fmt.Printf("⚠️  Final Risk Level: %s\n", ultimate.RiskLevel)

	// Example 24: Advanced Trading Decision Framework
	fmt.Println("\n=== 🤖 AI BOT DECISION FRAMEWORK 🤖 ===")

	switch ultimate.FinalSignal {
	case "STRONG BUY":
		fmt.Println("🚀 EXECUTE AGGRESSIVE BUY")
		fmt.Println("   ✅ All technical indicators bullish")
		fmt.Println("   ✅ Volume confirms breakout/accumulation")
		fmt.Println("   ✅ Low rug pull risk")
		fmt.Printf("   📊 Position: 3-5%% of portfolio (Risk: %s)\n", ultimate.RiskLevel)

	case "BUY":
		fmt.Println("📈 EXECUTE STANDARD BUY")
		fmt.Println("   ✅ Majority indicators bullish")
		if ultimate.VolumeConfirm {
			fmt.Println("   ✅ Volume supports the move")
		} else {
			fmt.Println("   ⚠️ Volume doesn't fully confirm")
		}
		fmt.Printf("   📊 Position: 2-3%% of portfolio (Risk: %s)\n", ultimate.RiskLevel)

	case "STRONG SELL":
		fmt.Println("🔴 EXECUTE IMMEDIATE SELL")
		fmt.Println("   ❌ All indicators bearish")
		fmt.Println("   ❌ High distribution detected")
		fmt.Printf("   🚨 Rug Pull Risk: %s\n", ultimate.RugPullRisk)

	case "SELL":
		fmt.Println("📉 EXECUTE GRADUAL SELL")
		fmt.Println("   ❌ Majority indicators bearish")
		fmt.Printf("   🚨 Rug Pull Risk: %s\n", ultimate.RugPullRisk)

	case "WAIT":
		fmt.Println("⏳ WAIT FOR OPTIMAL ENTRY")
		fmt.Println("   🔄 Low volatility squeeze detected")
		fmt.Println("   📊 Prepare for potential breakout")
		fmt.Println("   🔔 Set alerts for volume spikes")

	case "SUSPICIOUS":
		fmt.Println("🚨 SUSPICIOUS ACTIVITY DETECTED")
		fmt.Println("   ⚠️ Low volume on price moves")
		fmt.Println("   🤖 Potential bot manipulation")
		fmt.Println("   🚫 AVOID TRADING")

	case "HOLD":
		fmt.Println("🤔 MAINTAIN CURRENT POSITION")
		fmt.Println("   📊 Mixed or weak signals")
		fmt.Printf("   📈 Volume Confirmation: %v\n", ultimate.VolumeConfirm)
	}

	// Example 25: Risk Management with Volume
	fmt.Println("\n=== 🛡️ ADVANCED RISK MANAGEMENT 🛡️ ===")

	fmt.Printf("Rug Pull Risk Assessment: %s\n", ultimate.RugPullRisk)
	switch ultimate.RugPullRisk {
	case "extreme":
		fmt.Println("🚨 EXTREME DANGER - EXIT IMMEDIATELY")
		fmt.Println("   • Mass distribution detected")
		fmt.Println("   • High volume selling pressure")
		fmt.Println("   • Technical breakdown confirmed")

	case "high":
		fmt.Println("⚠️ HIGH RISK - REDUCE EXPOSURE")
		fmt.Println("   • Strong distribution signals")
		fmt.Println("   • Consider taking profits")
		fmt.Println("   • Tighten stop losses")

	case "medium":
		fmt.Println("⚡ MEDIUM RISK - STAY ALERT")
		fmt.Println("   • Some distribution detected")
		fmt.Println("   • Monitor volume closely")
		fmt.Println("   • Prepare exit strategy")

	case "low":
		fmt.Println("✅ LOW RISK - NORMAL OPERATION")
		fmt.Println("   • Healthy volume patterns")
		fmt.Println("   • No distribution signals")
		fmt.Println("   • Safe to hold/accumulate")
	}

	// Example 26: Volume-Based Entry/Exit Rules
	fmt.Println("\n=== 📋 VOLUME-BASED TRADING RULES 📋 ===")

	volumeRatio := ultimate.Volume.VolumeRatio
	fmt.Printf("Current Volume Ratio: %.2f\n", volumeRatio)

	switch {
	case volumeRatio >= 5.0:
		fmt.Println("🔥 EXTREME VOLUME - Major event likely")
		fmt.Println("   • Check news/announcements")
		fmt.Println("   • Prepare for high volatility")

	case volumeRatio >= 3.0:
		fmt.Println("📈 HIGH VOLUME - Strong interest")
		fmt.Println("   • Confirm with price action")
		fmt.Println("   • Good for breakout trades")

	case volumeRatio >= 1.5:
		fmt.Println("📊 ABOVE AVERAGE - Normal activity")
		fmt.Println("   • Standard trading conditions")

	case volumeRatio < 0.5:
		fmt.Println("⚠️ LOW VOLUME - Weak conviction")
		fmt.Println("   • Avoid trading")
		fmt.Println("   • Potential fake moves")
	}

}
