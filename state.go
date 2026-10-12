package main

type State struct {
	PriceAboveCloud bool
	MACDPositive    bool
	MACDUp          bool
	StochOverbought bool
	KeltnerUpperHit bool
}
