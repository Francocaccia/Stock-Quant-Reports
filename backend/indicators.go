package main

import "math"

type Indicators struct {
	Tenkan []float64
	Kijun  []float64
	SpanA  []float64
	SpanB  []float64

	KeltMid []float64
	KeltUp  []float64
	KeltDn  []float64

	MACD   []float64
	Signal []float64
	Hist   []float64

	StochK []float64
	StochD []float64

	Correlation []float64

	Volume []float64
}

func ComputeAllIndicators(bars []Bar, spyBars []Bar) Indicators {
	n := len(bars)

	tenkan := make([]float64, n)
	kijun := make([]float64, n)
	spanA := make([]float64, n)
	spanB := make([]float64, n)

	for i := 0; i < n; i++ {
		tenkan[i] = midHighLow(bars, i, 9)
		kijun[i] = midHighLow(bars, i, 26)
		spanA[i] = (tenkan[i] + kijun[i]) / 2.0
		spanB[i] = midHighLow(bars, i, 52)
	}

	close := make([]float64, n)
	high := make([]float64, n)
	low := make([]float64, n)
	vol := make([]float64, n)

	for i := range bars {
		close[i] = bars[i].C
		high[i] = bars[i].H
		low[i] = bars[i].L
		vol[i] = bars[i].V
	}

	spyClose := make([]float64, n)
	if len(spyBars) > 0 {
		spyIdx := 0
		for i := 0; i < n; i++ {
			t := bars[i].T
			// Advance to the SPY bar closest in time
			for spyIdx < len(spyBars)-1 {
				diffCurrent := math.Abs(float64(spyBars[spyIdx].T.Unix() - t.Unix()))
				diffNext := math.Abs(float64(spyBars[spyIdx+1].T.Unix() - t.Unix()))
				if diffNext <= diffCurrent {
					spyIdx++
				} else {
					break
				}
			}
			spyClose[i] = spyBars[spyIdx].C
		}
	}

	// KELTNER (SMA 20 + 1.5 ATR 20)
	sma20 := SMA(close, 20)
	tr := TrueRange(high, low, close)
	atr20 := SMA(tr, 20)

	kMid := sma20
	kUp := make([]float64, n)
	kDn := make([]float64, n)

	for i := 0; i < n; i++ {
		kUp[i] = kMid[i] + 1.5*atr20[i]
		kDn[i] = kMid[i] - 1.5*atr20[i]
	}

	// MACD
	macd := MACDLine(close, 12, 26)
	sig := EMA(macd, 9)
	hist := make([]float64, n)
	for i := 0; i < n; i++ {
		hist[i] = macd[i] - sig[i]
	}

	// STOCH (10,3)
	stochK := StochK(high, low, close, 10)
	stochD := SMA(stochK, 3)

	// CORRELATION 10
	corr := RollingCorrelation(close, spyClose, 10)

	return Indicators{
		Tenkan: tenkan,
		Kijun:  kijun,
		SpanA:  spanA,
		SpanB:  spanB,

		KeltMid: kMid,
		KeltUp:  kUp,
		KeltDn:  kDn,

		MACD:   macd,
		Signal: sig,
		Hist:   hist,

		StochK: stochK,
		StochD: stochD,

		Correlation: corr,

		Volume: vol,
	}
}

func midHighLow(bars []Bar, i int, lookback int) float64 {
	start := i - (lookback - 1)
	if start < 0 {
		start = 0
	}
	hi := bars[start].H
	lo := bars[start].L

	for j := start; j <= i; j++ {
		if bars[j].H > hi {
			hi = bars[j].H
		}
		if bars[j].L < lo {
			lo = bars[j].L
		}
	}
	return (hi + lo) / 2.0
}

func SMA(x []float64, period int) []float64 {
	n := len(x)
	out := make([]float64, n)

	for i := 0; i < n; i++ {
		start := i - (period - 1)
		if start < 0 {
			start = 0
		}
		sum := 0.0
		count := 0
		for j := start; j <= i; j++ {
			sum += x[j]
			count++
		}
		out[i] = sum / float64(count)
	}
	return out
}

func EMA(x []float64, period int) []float64 {
	n := len(x)
	out := make([]float64, n)
	if n == 0 {
		return out
	}
	alpha := 2.0 / float64(period+1)
	out[0] = x[0]
	for i := 1; i < n; i++ {
		out[i] = alpha*x[i] + (1-alpha)*out[i-1]
	}
	return out
}

func TrueRange(high, low, close []float64) []float64 {
	n := len(close)
	tr := make([]float64, n)

	for i := 0; i < n; i++ {
		if i == 0 {
			tr[i] = high[i] - low[i]
			continue
		}
		a := high[i] - low[i]
		b := math.Abs(high[i] - close[i-1])
		c := math.Abs(low[i] - close[i-1])
		tr[i] = max3(a, b, c)
	}
	return tr
}

func MACDLine(close []float64, fast, slow int) []float64 {
	ef := EMA(close, fast)
	es := EMA(close, slow)
	n := len(close)
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		out[i] = ef[i] - es[i]
	}
	return out
}

func StochK(high, low, close []float64, period int) []float64 {
	n := len(close)
	out := make([]float64, n)

	for i := 0; i < n; i++ {
		start := i - (period - 1)
		if start < 0 {
			start = 0
		}
		hi := high[start]
		lo := low[start]

		for j := start; j <= i; j++ {
			if high[j] > hi {
				hi = high[j]
			}
			if low[j] < lo {
				lo = low[j]
			}
		}

		den := hi - lo
		if den == 0 {
			out[i] = 50
		} else {
			out[i] = 100 * (close[i] - lo) / den
		}
	}
	return out
}

func RollingCorrelation(x, y []float64, period int) []float64 {
	n := len(x)
	out := make([]float64, n)

	for i := 0; i < n; i++ {
		start := i - (period - 1)
		if start < 0 {
			out[i] = 0
			continue
		}

		var sumX, sumY, sumXY, sumX2, sumY2 float64
		for j := start; j <= i; j++ {
			sumX += x[j]
			sumY += y[j]
			sumXY += x[j] * y[j]
			sumX2 += x[j] * x[j]
			sumY2 += y[j] * y[j]
		}

		nf := float64(period)
		num := nf*sumXY - sumX*sumY
		den := math.Sqrt((nf*sumX2 - sumX*sumX) * (nf*sumY2 - sumY*sumY))

		if den == 0 {
			out[i] = 0
		} else {
			out[i] = num / den
		}
	}

	return out
}


func FilterIndicators(ind Indicators, totalLen, filteredLen int) Indicators {
	start := totalLen - filteredLen
	if start < 0 {
		start = 0
	}

	return Indicators{
		Tenkan: ind.Tenkan[start:],
		Kijun:  ind.Kijun[start:],
		SpanA:  ind.SpanA[start:],
		SpanB:  ind.SpanB[start:],

		KeltMid: ind.KeltMid[start:],
		KeltUp:  ind.KeltUp[start:],
		KeltDn:  ind.KeltDn[start:],

		MACD:   ind.MACD[start:],
		Signal: ind.Signal[start:],
		Hist:   ind.Hist[start:],

		StochK: ind.StochK[start:],
		StochD: ind.StochD[start:],

		Correlation: ind.Correlation[start:],

		Volume: ind.Volume[start:],
	}
}

func max3(a, b, c float64) float64 {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}