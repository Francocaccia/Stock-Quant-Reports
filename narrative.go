package main

import "fmt"

// WeeklyNarrative produces a tight 4-6 sentence analytical paragraph
// that cross-references indicators and includes specific price levels.
func WeeklyNarrative(s State, ind Indicators, bars []Bar) string {
	n := len(bars)
	if n == 0 {
		return ""
	}

	price := bars[n-1].C
	kijun := ind.Kijun[n-1]
	keltUp := ind.KeltUp[n-1]
	keltDn := ind.KeltDn[n-1]

	// Detect MACD acceleration vs deceleration
	macdAccel := false
	if n >= 2 {
		macdAccel = ind.MACD[n-1] > ind.MACD[n-2]
	}

	// Stochastic K vs D crossover direction
	stochKAboveD := ind.StochK[n-1] > ind.StochD[n-1]

	var text string

	// 1. Opening — price vs cloud + Kijun
	if s.PriceAboveCloud {
		text += fmt.Sprintf(
			"On the weekly timeframe, price (%.2f) remains above the Ichimoku cloud with the Kijun at %.2f acting as near-term support. ",
			price, kijun)
	} else {
		text += fmt.Sprintf(
			"On the weekly timeframe, price (%.2f) has broken below the Ichimoku cloud, with the Kijun at %.2f now acting as resistance — a structural deterioration. ",
			price, kijun)
	}

	// 2. MACD — cross-referenced with trend
	if s.MACDPositive && macdAccel {
		text += "The MACD is positive and accelerating, confirming that bullish momentum is expanding and support the prevailing uptrend. "
	} else if s.MACDPositive && !macdAccel {
		if s.PriceAboveCloud {
			text += "The MACD remains positive but is rolling over — momentum is diverging from structure, trend intact but losing conviction. "
		} else {
			text += "The MACD is positive but decelerating, offering limited confirmation for a recovery while price trades below cloud support. "
		}
	} else {
		text += "The MACD is negative, indicating that downside momentum dominates and reinforces the bearish structural bias. "
	}

	// 3. Stochastic — cross-referenced with MACD
	if s.StochOverbought && !macdAccel {
		text += "Stochastic readings are overbought (K above 80) while momentum is fading — a bearish divergence that raises near-term pullback risk. "
	} else if s.StochOverbought {
		text += "Stochastic is elevated but K remains above D, suggesting the momentum push is not yet exhausted despite overbought conditions. "
	} else if stochKAboveD {
		if s.PriceAboveCloud && s.MACDPositive {
			text += "Stochastic is in mid-range with K above D, indicating orderly upside participation with room to extend before becoming stretched. "
		} else if s.PriceAboveCloud && !s.MACDPositive {
			text += "Stochastic is in mid-range with K above D, reflecting a short-term bounce in momentum within a weakening trend structure. "
		} else {
			text += "Stochastic is in mid-range with K above D, reflecting a short-term bounce within a bearish structure, insufficient to alter the dominant downtrend. "
		}
	} else {
		if !s.PriceAboveCloud {
			text += "Stochastic K has crossed below D, with momentum deteriorating, consistent with the prevailing downside bias. "
		} else {
			text += "Stochastic K has crossed below D, adding a short-term cautionary signal even if the broader trend has not yet reversed. "
		}
	}

	// 4. Keltner with specific levels
	if s.KeltnerUpperHit {
		text += fmt.Sprintf(
			"Price is pressing the upper Keltner band (%.2f), a level historically associated with mean-reversion risk — positioning extended relative to the channel envelope. ",
			keltUp)
	} else {
		text += fmt.Sprintf(
			"Price sits within the Keltner Channel (%.2f – %.2f), with sufficient room to advance before reaching extended territory. ",
			keltDn, keltUp)
	}

	return text
}

// MonthlyNarrative produces a tight 4-6 sentence long-term analytical paragraph.
func MonthlyNarrative(s State, ind Indicators, bars []Bar, weeklyS State) string {
	n := len(bars)
	if n == 0 {
		return ""
	}

	price := bars[n-1].C
	kijun := ind.Kijun[n-1]
	keltUp := ind.KeltUp[n-1]
	keltDn := ind.KeltDn[n-1]

	macdAccel := false
	if n >= 2 {
		macdAccel = ind.MACD[n-1] > ind.MACD[n-2]
	}

	stochKAboveD := ind.StochK[n-1] > ind.StochD[n-1]

	// Detect weekly vs monthly divergence
	weeklyMonthlyConflict := weeklyS.PriceAboveCloud != s.PriceAboveCloud ||
		weeklyS.MACDPositive != s.MACDPositive

	var text string

	// 1. Opening — long-term structural read
	if s.PriceAboveCloud {
		text += fmt.Sprintf(
			"On the monthly timeframe, the structure remains strongly bullish — price (%.2f) is well above the cloud with the Kijun at %.2f, confirming long-term trend integrity. ",
			price, kijun)
	} else {
		text += fmt.Sprintf(
			"On the monthly timeframe, price (%.2f) remains below the cloud with the Kijun at %.2f providing overhead resistance, signaling a structurally bearish long-term regime. ",
			price, kijun)
	}

	// 2. MACD at macro level
	if !s.PriceAboveCloud && s.MACDPositive {
		text += "The MACD has turned positive from depressed levels, reflecting a potential momentum inflection, though structural resistance from the cloud overhead limits conviction. "
	} else if s.MACDPositive && macdAccel {
		text += "The MACD remains positive and expanding, supporting continuation with strong long-term momentum behind the move. "
	} else if s.MACDPositive {
		text += "The MACD remains positive, supporting continuation, but the rate of expansion is slowing — the bull cycle may be maturing. "
	} else {
		text += "Monthly MACD has turned negative, a significant long-term deterioration that limits conviction for any recovery attempt. "
	}

	// 3. Stochastic at macro level
	if s.StochOverbought {
		text += "The Stochastic is elevated and flattening, suggesting the trend is mature and consolidation is more likely than immediate continuation. "
	} else if stochKAboveD {
		if s.PriceAboveCloud && s.MACDPositive {
			text += "Monthly Stochastic shows K above D in mid-range, indicating orderly upside participation with room to extend before becoming stretched. "
		} else if s.PriceAboveCloud && !s.MACDPositive {
			text += "Monthly Stochastic shows K above D in mid-range, reflecting a short-term bounce in momentum within a weakening trend structure. "
		} else {
			text += "Monthly Stochastic shows K above D in mid-range, reflecting a short-term bounce within a bearish structure, insufficient to alter the dominant downtrend. "
		}
	} else {
		if !s.PriceAboveCloud {
			text += "Stochastic K has crossed below D on the monthly chart, with momentum deteriorating, consistent with the prevailing downside bias. "
		} else {
			text += "Stochastic K has crossed below D on the monthly chart — a longer-cycle warning that warrants attention even if price structure holds. "
		}
	}

	// 4. Keltner at macro scale
	if s.KeltnerUpperHit {
		text += fmt.Sprintf(
			"Price has reached the upper Keltner band (%.2f) on the monthly chart, a level historically associated with long-term extended positioning and mean-reversion risk. ",
			keltUp)
	} else {
		text += fmt.Sprintf(
			"Price remains within the Keltner Channel (%.2f – %.2f), consistent with an orderly trend without signs of parabolic extension. ",
			keltDn, keltUp)
	}

	// 5. Weekly/Monthly divergence
	if weeklyMonthlyConflict {
		if s.PriceAboveCloud && !weeklyS.PriceAboveCloud {
			text += "Weekly structure shows deterioration while monthly remains constructive — a tactical weakness within a strategic uptrend, favoring caution on new entries but no structural exit."
		} else if !s.PriceAboveCloud && weeklyS.PriceAboveCloud {
			text += "Monthly structure is bearish while weekly shows a bounce — a tactical recovery within a strategic downtrend; rallies should be treated as selling opportunities."
		}
	}

	return text
}