package main

func BuildState(bars []Bar, ind Indicators) State {
	n := len(bars)
	if n == 0 {
		return State{}
	}
	i := n - 1

	price := bars[i].C

	// Cloud Offset: Standard Ichimoku compares price with Cloud from 26 periods ago
	lookback := 26
	cloudIdx := i - lookback
	if cloudIdx < 0 {
		cloudIdx = 0
	}

	top := max(ind.SpanA[cloudIdx], ind.SpanB[cloudIdx])

	macdPos := ind.MACD[i] > 0
	macdUp := false
	if i >= 2 {
		macdUp = ind.MACD[i] > ind.MACD[i-1]
	}

	stochOB := ind.StochK[i] >= 80
	keltHit := price >= ind.KeltUp[i]

	return State{
		PriceAboveCloud: price > top,
		MACDPositive:    macdPos,
		MACDUp:          macdUp,
		StochOverbought: stochOB,
		KeltnerUpperHit: keltHit,
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
