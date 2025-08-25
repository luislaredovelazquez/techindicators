package techindicators

import (
	"fmt"
	"strconv"
	"time"
)

// ExampleUsage demonstrates comprehensive usage of all technical indicators with OHLCV data
func ExampleUsage() {
	// Create sample OHLCV data for demonstration
	dataset := []OHLCV{
		{Timestamp: time.Unix(1692230400, 0), Open: 0.001, Close: 0.0012, High: 0.0013, Low: 0.0009, Volume: 1000000},
		{Timestamp: time.Unix(1692234000, 0), Open: 0.0012, Close: 0.0011, High: 0.0014, Low: 0.001, Volume: 1200000},
		{Timestamp: time.Unix(1692237600, 0), Open: 0.0011, Close: 0.0013, High: 0.0015, Low: 0.001, Volume: 900000},
		{Timestamp: time.Unix(1692241200, 0), Open: 0.0013, Close: 0.0014, High: 0.0016, Low: 0.0012, Volume: 1100000},
		{Timestamp: time.Unix(1692244800, 0), Open: 0.0014, Close: 0.0012, High: 0.0015, Low: 0.0011, Volume: 800000},
		{Timestamp: time.Unix(1692248400, 0), Open: 0.0012, Close: 0.0015, High: 0.0016, Low: 0.0011, Volume: 1300000},
		{Timestamp: time.Unix(1692252000, 0), Open: 0.0015, Close: 0.0016, High: 0.0018, Low: 0.0014, Volume: 1500000},
		{Timestamp: time.Unix(1692255600, 0), Open: 0.0016, Close: 0.0014, High: 0.0017, Low: 0.0013, Volume: 1100000},
		{Timestamp: time.Unix(1692259200, 0), Open: 0.0014, Close: 0.0017, High: 0.0019, Low: 0.0013, Volume: 1400000},
		{Timestamp: time.Unix(1692262800, 0), Open: 0.0017, Close: 0.0018, High: 0.002, Low: 0.0016, Volume: 1600000},
		{Timestamp: time.Unix(1692266400, 0), Open: 0.0018, Close: 0.0019, High: 0.0021, Low: 0.0017, Volume: 1700000},
		{Timestamp: time.Unix(1692270000, 0), Open: 0.0019, Close: 0.002, High: 0.0022, Low: 0.0018, Volume: 1800000},
		{Timestamp: time.Unix(1692273600, 0), Open: 0.002, Close: 0.0018, High: 0.0021, Low: 0.0017, Volume: 1200000},
		{Timestamp: time.Unix(1692277200, 0), Open: 0.0018, Close: 0.0022, High: 0.0023, Low: 0.0017, Volume: 2000000},
		{Timestamp: time.Unix(1692280800, 0), Open: 0.0022, Close: 0.0021, High: 0.0024, Low: 0.002, Volume: 1600000},
	}

	fmt.Println("=== TECHNICAL INDICATORS EXAMPLE USAGE ===")
	fmt.Printf("Dataset contains %d OHLCV candles\n\n", len(dataset))

	// ===================
	// SMA Analysis
	// ===================
	fmt.Println("=== SMA ANALYSIS ===")

	// Calculate SMA-5 using close prices
	sma5, err := CalculateSMA(dataset, 5, ClosePrice)
	if err != nil {
		fmt.Printf("Error calculating SMA-5: %v\n", err)
		return
	}

	fmt.Println("SMA-5 (Close Prices) - Last 3 values:")
	for i := len(sma5) - 3; i < len(sma5); i++ {
		if i >= 0 {
			fmt.Printf("  %s: %.6f\n", sma5[i].Timestamp, sma5[i].Value)
		}
	}

	// Get latest SMA value
	latestSMA5, err := GetLatestSMA(dataset, 5, ClosePrice)
	if err != nil {
		fmt.Printf("Error getting latest SMA: %v\n", err)
		return
	}
	fmt.Printf("Latest SMA-5: %.6f\n", latestSMA5)

	// Check if price is above SMA (bullish signal)
	isAbove, err := IsPriceAboveSMA(dataset, 5, ClosePrice)
	if err != nil {
		fmt.Printf("Error checking price vs SMA: %v\n", err)
		return
	}
	fmt.Printf("Is current price above SMA-5? %v\n", isAbove)

	// SMA Crossover detection
	crossover, err := SMACrossover(dataset, 3, 5, ClosePrice)
	if err != nil {
		fmt.Printf("Error detecting crossover: %v\n", err)
		return
	}
	fmt.Printf("SMA Crossover Signal (3 vs 5): %s\n\n", crossover)

	// ===================
	// Bollinger Bands Analysis
	// ===================
	fmt.Println("=== BOLLINGER BANDS ANALYSIS ===")

	// Calculate Bollinger Bands
	bb, err := CalculateBollingerBands(dataset, 5, 2.0, ClosePrice)
	if err != nil {
		fmt.Printf("Error calculating Bollinger Bands: %v\n", err)
		return
	}

	fmt.Println("Bollinger Bands (5 period, 2.0 multiplier) - Last 2 values:")
	for i := len(bb) - 2; i < len(bb); i++ {
		if i >= 0 {
			fmt.Printf("  %s:\n", bb[i].Timestamp)
			fmt.Printf("    Upper: %.6f, Middle: %.6f, Lower: %.6f\n", bb[i].UpperBand, bb[i].MiddleBand, bb[i].LowerBand)
			fmt.Printf("    Band Width: %.4f\n", bb[i].BandWidth)
		}
	}

	// Get current price position relative to bands
	position, err := GetPricePosition(dataset, 5, 2.0, ClosePrice, 0.02)
	if err != nil {
		fmt.Printf("Error getting price position: %v\n", err)
		return
	}
	fmt.Printf("Current Price Position: %s\n", position)

	// Complete Bollinger Strategy Analysis
	bbStrategy, err := AnalyzeBollingerStrategy(dataset, 5, 2.0, ClosePrice)
	if err != nil {
		fmt.Printf("Error analyzing Bollinger strategy: %v\n", err)
		return
	}
	fmt.Printf("Bollinger Strategy Signal: %s\n\n", bbStrategy.Signal)

	// ===================
	// RSI Analysis
	// ===================
	fmt.Println("=== RSI ANALYSIS ===")

	// Calculate RSI
	rsi, err := CalculateRSI(dataset, 5, ClosePrice)
	if err != nil {
		fmt.Printf("Error calculating RSI: %v\n", err)
		return
	}

	fmt.Println("RSI (5 period) - Last 3 values:")
	for i := len(rsi) - 3; i < len(rsi); i++ {
		if i >= 0 {
			fmt.Printf("  %s: %.2f (%s)\n", rsi[i].Timestamp, rsi[i].Value, rsi[i].Signal)
		}
	}

	// Complete RSI Strategy
	rsiStrategy, err := AnalyzeRSIStrategy(dataset, 5, ClosePrice)
	if err != nil {
		fmt.Printf("Error analyzing RSI strategy: %v\n", err)
		return
	}
	fmt.Printf("RSI Strategy Signal: %s (Condition: %s)\n\n", rsiStrategy.Signal, rsiStrategy.Condition)

	// ===================
	// Volume Analysis
	// ===================
	fmt.Println("=== VOLUME ANALYSIS ===")

	// Calculate volume indicators
	volumeResults, err := CalculateVolumeAnalysis(dataset, 5, 3)
	if err != nil {
		fmt.Printf("Error calculating volume analysis: %v\n", err)
		return
	}

	fmt.Println("Volume Analysis - Last 2 values:")
	for i := len(volumeResults) - 2; i < len(volumeResults); i++ {
		if i >= 0 {
			vol := volumeResults[i]
			fmt.Printf("  %s:\n", vol.Timestamp)
			fmt.Printf("    Volume: %.0f, VMA: %.0f, Ratio: %.2f\n", vol.Volume, vol.VMA, vol.Volume/vol.VMA)
			fmt.Printf("    OBV: %.0f, VPT: %.2f, ADL: %.2f\n", vol.OBV, vol.VPT, vol.ADL)
		}
	}

	// Complete Volume Strategy
	volumeStrategy, err := AnalyzeVolumeStrategy(dataset, 5, 3)
	if err != nil {
		fmt.Printf("Error analyzing volume strategy: %v\n", err)
		return
	}
	fmt.Printf("Volume Strategy Signal: %s (OBV Trend: %s)\n\n", volumeStrategy.Signal, volumeStrategy.OBVTrend)

	// ===================
	// Comprehensive Analysis
	// ===================
	fmt.Println("=== COMPREHENSIVE TECHNICAL ANALYSIS ===")

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
	fmt.Printf("⚠️  Risk Level: %s\n\n", comprehensive.RiskLevel)

	// ===================
	// Ultimate Analysis
	// ===================
	fmt.Println("=== 🚀 ULTIMATE MEMECOIN ANALYSIS 🚀 ===")

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
	fmt.Printf("⚠️  Final Risk Level: %s\n\n", ultimate.RiskLevel)

	// ===================
	// Trading Recommendations
	// ===================
	fmt.Println("=== 💰 TRADING RECOMMENDATIONS 💰 ===")

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

	default:
		fmt.Println("🤔 MAINTAIN CURRENT POSITION")
		fmt.Println("   📊 Mixed or weak signals")
		fmt.Printf("   📈 Volume Confirmation: %v\n", ultimate.VolumeConfirm)
	}

	fmt.Println("\n💡 Remember: This is a demonstration with limited data.")
	fmt.Println("💡 In live trading, use at least 100+ candles for reliable signals!")
	fmt.Println("💡 Always combine with fundamental analysis and market sentiment!")
}

// ConvertStringDataToOHLCV converts old [][]string format to new OHLCV format
// This helper function can be used to migrate existing data
func ConvertStringDataToOHLCV(stringData [][]string) ([]OHLCV, error) {
	if len(stringData) == 0 {
		return nil, fmt.Errorf("empty dataset")
	}

	var ohlcvData []OHLCV

	for i, candle := range stringData {
		if len(candle) < 6 {
			return nil, fmt.Errorf("invalid candle at index %d: expected 6 fields, got %d", i, len(candle))
		}

		// Parse timestamp (assuming Unix timestamp)
		var timestamp time.Time
		if unixTime, err := parseFloat64(candle[0]); err == nil {
			timestamp = time.Unix(int64(unixTime), 0)
		} else {
			// If not Unix timestamp, try parsing as RFC3339
			if t, err := time.Parse(time.RFC3339, candle[0]); err == nil {
				timestamp = t
			} else {
				return nil, fmt.Errorf("invalid timestamp at index %d: %s", i, candle[0])
			}
		}

		open, err := parseFloat64(candle[1])
		if err != nil {
			return nil, fmt.Errorf("invalid open price at index %d: %w", i, err)
		}

		close, err := parseFloat64(candle[2])
		if err != nil {
			return nil, fmt.Errorf("invalid close price at index %d: %w", i, err)
		}

		high, err := parseFloat64(candle[3])
		if err != nil {
			return nil, fmt.Errorf("invalid high price at index %d: %w", i, err)
		}

		low, err := parseFloat64(candle[4])
		if err != nil {
			return nil, fmt.Errorf("invalid low price at index %d: %w", i, err)
		}

		volume, err := parseFloat64(candle[5])
		if err != nil {
			return nil, fmt.Errorf("invalid volume at index %d: %w", i, err)
		}

		ohlcvData = append(ohlcvData, OHLCV{
			Timestamp: timestamp,
			Open:      open,
			Close:     close,
			High:      high,
			Low:       low,
			Volume:    volume,
		})
	}

	return ohlcvData, nil
}

// parseFloat64 is a helper function to parse string to float64
func parseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
