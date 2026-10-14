package main

func GenerateSignal(weekly State, monthly State) (string, string) {

	score := 0

	// Structural conditions (1 pt each)
	if weekly.PriceAboveCloud {
		score++
	}
	if monthly.PriceAboveCloud {
		score++
	}

	// Momentum conditions (1 pt each)
	if weekly.MACDPositive && weekly.PriceAboveCloud {
		score++
	}
	if monthly.MACDPositive && monthly.PriceAboveCloud {
		score++
	}

	// Momentum acceleration bonus (1 pt each)
	if weekly.MACDUp && weekly.PriceAboveCloud {
		score++
	}
	if monthly.MACDUp && monthly.PriceAboveCloud {
		score++
	}

	// Max score = 6
	var signal, text string

	switch {
	case score >= 5:
		signal = "BUY (HIGH CONVICTION)"
		text = "The asset exhibits strong multi-timeframe alignment, with structural integrity and momentum confirmation — and acceleration — across both weekly and monthly horizons. Trend configuration favors sustained continuation of the prevailing bullish regime."

	case score == 4:
		signal = "BUY (HIGH CONVICTION)"
		text = "Multi-timeframe structure is constructive with positive momentum across both horizons. Minor acceleration signals are incomplete, but the overall configuration strongly favors continuation."

	case score == 3:
		signal = "BUY (MODERATE)"
		text = "Technical conditions are broadly constructive but with divergences limiting full confirmation. The asset favors continuation with moderated conviction — size positions accordingly."

	case score == 2:
		signal = "HOLD"
		text = "Mixed technical signals across weekly and monthly timeframes suggest consolidation is the highest-probability near-term outcome. The configuration does not support aggressive positioning in either direction."

	case score == 1:
		signal = "SELL (MODERATE)"
		text = "Deteriorating conditions across key technical dimensions increase downside risk. The weight of evidence favors reducing exposure and awaiting structural recovery before re-engaging."

	default:
		signal = "SELL (HIGH CONVICTION)"
		text = "Structural deterioration and negative momentum across all tracked dimensions signal elevated downside risk. Technical configuration strongly favors defensive allocation over active market participation."
	}

	return signal, text
}