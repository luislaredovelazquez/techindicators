# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library for financial technical indicators (`github.com/luislaredovelazquez/techindicators`) that provides comprehensive analysis tools for trading, particularly focused on cryptocurrency and memecoin analysis. The library implements various technical indicators and combines them for sophisticated trading signal generation.

## Core Architecture

### Technical Indicators
The library is structured around modular technical indicators, each in separate files:

- **Simple Moving Average (SMA)** - `movingAvg.go`: Trend analysis with crossover detection
- **Bollinger Bands** - `bollingerBands.go`: Volatility-based analysis with squeeze detection
- **Relative Strength Index (RSI)** - `rsi.go`: Momentum oscillator with divergence detection
- **Volume Analysis** - `volumeAnalysis.go`: Comprehensive volume indicators (VMA, OBV, VPT, VROC, ADL)
- **Sharpe Ratio** - `sharpeRatio.go`: Risk-adjusted return calculation using CoinGecko API

### Data Structure
All indicators expect candle data in the format: `[][]string` where each candle is:
```
[timestamp, open, close, high, low, volume]
```

### Price Types
The library supports multiple price extraction methods via `PriceType`:
- `ClosePrice`, `OpenPrice`, `HighPrice`, `LowPrice`
- `TypicalPrice` (H+L+C)/3
- `WeightedPrice` (H+L+2*C)/4

### Signal Generation
The library provides three levels of analysis:
1. **Individual indicators** - Basic buy/sell signals
2. **ComprehensiveAnalysis** - Combines SMA, Bollinger Bands, and RSI
3. **UltimateAnalysis** - Adds volume confirmation and rug pull risk assessment

## Development Commands

### Building and Testing
```bash
# Build the module
go build -v .

# Format code
go fmt .

# Run static analysis
go vet .

# Tidy dependencies
go mod tidy

# Test (currently no test files exist)
go test -v .
```

### Dependencies
- `github.com/JulianToledano/goingecko/v3` - CoinGecko API client
- `github.com/mark3labs/mcp-go` - MCP (Model Context Protocol) framework

## Key Functions and Components

### Entry Points
- `ComprehensiveAnalysis()` - Main technical analysis combining all indicators
- `UltimateAnalysis()` - Complete analysis including volume and risk assessment
- `SharpeRatioHandler()` - MCP tool handler for Sharpe ratio calculation

### Strategy Components
- **BollingerStrategy** - Position detection, breakout signals, squeeze analysis
- **RSIStrategy** - Momentum analysis with divergence detection
- **VolumeStrategy** - Breakout detection, accumulation/distribution analysis
- **UltimateMemecoinAnalysis** - Complete trading framework with rug pull detection

### Critical Signal Types
The library generates various signal strengths:
- Technical signals: `STRONG BUY`, `BUY`, `HOLD`, `SELL`, `STRONG SELL`, `WAIT`, `SUSPICIOUS`
- Volume signals: `strong_buy`, `buy`, `accumulate`, `distribute`, `sell`, `strong_sell`, `low_volume_alert`
- Risk levels: `LOW`, `MEDIUM`, `HIGH`
- Rug pull risk: `low`, `medium`, `high`, `extreme`

## Trading Logic

The library implements sophisticated trading logic that:
1. Analyzes individual indicators for basic signals
2. Combines signals with weighted voting for final decisions
3. Uses volume confirmation to validate technical signals
4. Assesses rug pull risk based on distribution patterns
5. Adjusts confidence levels based on signal alignment

The `UltimateAnalysis` function provides the most comprehensive analysis, combining all indicators with volume confirmation and risk assessment specifically designed for memecoin trading.