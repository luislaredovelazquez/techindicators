package techindicators

import (
	"time"
)

// OHLCV represents a single candle with Open, High, Low, Close, Volume data
type OHLCV struct {
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
}

// ExtractPrice extracts the specified price type from OHLCV data
func (o OHLCV) ExtractPrice(priceType PriceType) float64 {
	switch priceType {
	case OpenPrice:
		return o.Open
	case ClosePrice:
		return o.Close
	case HighPrice:
		return o.High
	case LowPrice:
		return o.Low
	case TypicalPrice:
		return (o.High + o.Low + o.Close) / 3
	case WeightedPrice:
		return (o.High + o.Low + 2*o.Close) / 4
	default:
		return o.Close // Default to close price
	}
}
