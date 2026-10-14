package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ─── Colour palette ──────────────────────────────────────────────────────────
var (
	colBackground  = color.RGBA{10, 10, 10, 255}
	colGrid        = color.RGBA{26, 26, 26, 255}
	colAxis        = color.RGBA{60, 60, 60, 255}
	colText        = color.RGBA{180, 180, 180, 255}
	colPanelLabel  = color.RGBA{90, 90, 90, 255}
	colCandleUp    = color.RGBA{0, 200, 100, 255}
	colCandleDn    = color.RGBA{204, 34, 0, 255}
	colVolUp       = color.RGBA{0, 80, 40, 200}
	colVolDn       = color.RGBA{100, 20, 0, 200}
	colCloudBull   = color.RGBA{0, 60, 0, 100}
	colCloudBear   = color.RGBA{60, 0, 0, 100}
	colTenkan      = color.RGBA{255, 102, 0, 220}  // orange
	colKijun       = color.RGBA{0, 102, 255, 220}  // blue
	colKeltUp      = color.RGBA{0, 180, 80, 200}   // green
	colKeltDn      = color.RGBA{0, 180, 80, 200}
	colKeltMid     = color.RGBA{0, 120, 50, 160}
	colKeltFill    = color.RGBA{20, 50, 20, 60}
	colMACDLine    = color.RGBA{0, 200, 255, 255}  // cyan
	colSignalLine  = color.RGBA{255, 102, 0, 255}  // orange
	colHistPos     = color.RGBA{0, 200, 80, 255}
	colHistNeg     = color.RGBA{200, 40, 0, 255}
	colZero        = color.RGBA{60, 60, 60, 255}
	colStochK      = color.RGBA{255, 0, 255, 255}  // magenta
	colStochD      = color.RGBA{255, 255, 0, 255}  // yellow
	colOB          = color.RGBA{180, 0, 0, 200}
	colOS          = color.RGBA{0, 180, 0, 200}
	colOBFill      = color.RGBA{180, 0, 0, 30}
	colOSFill      = color.RGBA{0, 180, 0, 30}
	colCorr        = color.RGBA{255, 170, 0, 255}  // amber
)

const (
	imgW = 1400
	imgH = 900
)

// ─── Panel layout ────────────────────────────────────────────────────────────
// Price panel = 55%, MACD = 20%, Stoch = 17%, Corr = 8%
// Inside the price panel, the bottom 12% is integrated volume bars.
type panel struct {
	y0, y1 int // pixel rows (inclusive), top-to-bottom
}

func makePanels() (price, macd, stoch, corr panel) {
	marginT := 10
	marginB := 30
	usable := imgH - marginT - marginB

	price = panel{marginT, marginT + int(float64(usable)*0.55)}
	macd = panel{price.y1 + 1, price.y1 + 1 + int(float64(usable)*0.20)}
	stoch = panel{macd.y1 + 1, macd.y1 + 1 + int(float64(usable)*0.17)}
	corr = panel{stoch.y1 + 1, marginT + usable}
	return
}

// ─── Drawing primitives ──────────────────────────────────────────────────────

func setPixel(img *image.RGBA, x, y int, c color.RGBA) {
	if x < 0 || x >= imgW || y < 0 || y >= imgH {
		return
	}
	img.SetRGBA(x, y, c)
}

// blend alpha-composite src over dst
func blendPixel(img *image.RGBA, x, y int, c color.RGBA) {
	if x < 0 || x >= imgW || y < 0 || y >= imgH {
		return
	}
	dst := img.RGBAAt(x, y)
	a := float64(c.A) / 255.0
	r := uint8(float64(c.R)*a + float64(dst.R)*(1-a))
	g := uint8(float64(c.G)*a + float64(dst.G)*(1-a))
	b := uint8(float64(c.B)*a + float64(dst.B)*(1-a))
	img.SetRGBA(x, y, color.RGBA{r, g, b, 255})
}

func hLine(img *image.RGBA, x0, x1, y int, c color.RGBA) {
	for x := x0; x <= x1; x++ {
		setPixel(img, x, y, c)
	}
}

func vLine(img *image.RGBA, x, y0, y1 int, c color.RGBA) {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for y := y0; y <= y1; y++ {
		setPixel(img, x, y, c)
	}
}

func vLineBlend(img *image.RGBA, x, y0, y1 int, c color.RGBA) {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for y := y0; y <= y1; y++ {
		blendPixel(img, x, y, c)
	}
}

// thickHLine draws a horizontal line with thickness t
func thickHLine(img *image.RGBA, x0, x1, y, t int, c color.RGBA) {
	for dy := -t / 2; dy <= t/2; dy++ {
		hLine(img, x0, x1, y+dy, c)
	}
}

// drawLine Bresenham
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	for {
		setPixel(img, x0, y0, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// fillRect fills a rectangle
func fillRect(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			blendPixel(img, x, y, c)
		}
	}
}

// ─── Scaling helpers ─────────────────────────────────────────────────────────
const (
	marginL = 55 // px left margin for Y axis labels
	marginR = 10
)

func scaleX(i, n int) int {
	usable := imgW - marginL - marginR
	return marginL + i*usable/n
}

func scaleY(v, vMin, vMax float64, p panel) int {
	if vMax == vMin {
		return (p.y0 + p.y1) / 2
	}
	ratio := (vMax - v) / (vMax - vMin)
	return p.y0 + int(ratio*float64(p.y1-p.y0))
}

func minMaxSlice(s []float64) (float64, float64) {
	if len(s) == 0 {
		return 0, 1
	}
	mn, mx := s[0], s[0]
	for _, v := range s {
		if v < mn {
			mn = v
		}
		if v > mx {
			mx = v
		}
	}
	return mn, mx
}

// ─── Draw a series line ───────────────────────────────────────────────────────
func drawSeries(img *image.RGBA, vals []float64, vMin, vMax float64, p panel, c color.RGBA) {
	n := len(vals)
	for i := 1; i < n; i++ {
		x0 := scaleX(i-1, n)
		x1 := scaleX(i, n)
		y0 := scaleY(vals[i-1], vMin, vMax, p)
		y1 := scaleY(vals[i], vMin, vMax, p)
		drawLine(img, x0, y0, x1, y1, c)
	}
}

// ─── Panel background + grid ─────────────────────────────────────────────────
func drawPanelBG(img *image.RGBA, p panel, nGridLines int, vMin, vMax float64) {
	// Fill background
	fillRect(img, 0, p.y0, imgW-1, p.y1, colBackground)
	// Horizontal grid lines
	for i := 0; i <= nGridLines; i++ {
		y := p.y0 + i*(p.y1-p.y0)/nGridLines
		hLine(img, marginL, imgW-marginR, y, colGrid)
	}
	// Top border
	hLine(img, 0, imgW-1, p.y0, colAxis)
}

// ─── Draw X axis date labels (every ~10% of bars) ───────────────────────────
func drawDateLabels(img *image.RGBA, bars []Bar, p panel) {
	n := len(bars)
	step := n / 8
	if step < 1 {
		step = 1
	}
	for i := 0; i < n; i += step {
		x := scaleX(i, n)
		label := bars[i].T.Format("Jan '06")
		drawSimpleText(img, label, x-12, p.y1+14, colText)
	}
}

// drawSimpleText renders a very small bitmap font (3×5 pixels per char).
// This keeps us dependency-free — good enough for chart labels.
// Each character is encoded as 5 rows of 3 bits.
var font3x5 = map[rune][5]uint8{
	'0': {0b111, 0b101, 0b101, 0b101, 0b111},
	'1': {0b010, 0b110, 0b010, 0b010, 0b111},
	'2': {0b111, 0b001, 0b111, 0b100, 0b111},
	'3': {0b111, 0b001, 0b111, 0b001, 0b111},
	'4': {0b101, 0b101, 0b111, 0b001, 0b001},
	'5': {0b111, 0b100, 0b111, 0b001, 0b111},
	'6': {0b111, 0b100, 0b111, 0b101, 0b111},
	'7': {0b111, 0b001, 0b001, 0b001, 0b001},
	'8': {0b111, 0b101, 0b111, 0b101, 0b111},
	'9': {0b111, 0b101, 0b111, 0b001, 0b111},
	'J': {0b001, 0b001, 0b001, 0b101, 0b111},
	'F': {0b111, 0b100, 0b110, 0b100, 0b100},
	'M': {0b101, 0b111, 0b101, 0b101, 0b101},
	'A': {0b010, 0b101, 0b111, 0b101, 0b101},
	'S': {0b111, 0b100, 0b111, 0b001, 0b111},
	'O': {0b111, 0b101, 0b101, 0b101, 0b111},
	'N': {0b101, 0b111, 0b111, 0b101, 0b101},
	'D': {0b110, 0b101, 0b101, 0b101, 0b110},
	'\'': {0b011, 0b010, 0b000, 0b000, 0b000},
	'a': {0b000, 0b111, 0b101, 0b101, 0b111},
	'b': {0b100, 0b110, 0b101, 0b101, 0b111},
	'c': {0b000, 0b111, 0b100, 0b100, 0b111},
	'd': {0b001, 0b011, 0b101, 0b101, 0b111},
	'e': {0b000, 0b111, 0b111, 0b100, 0b111},
	'f': {0b011, 0b010, 0b111, 0b010, 0b010},
	'g': {0b000, 0b111, 0b101, 0b111, 0b001},
	'h': {0b100, 0b100, 0b110, 0b101, 0b101},
	'i': {0b010, 0b000, 0b010, 0b010, 0b010},
	'j': {0b001, 0b000, 0b001, 0b101, 0b111},
	'k': {0b100, 0b101, 0b110, 0b110, 0b101},
	'l': {0b110, 0b010, 0b010, 0b010, 0b111},
	'm': {0b000, 0b101, 0b111, 0b111, 0b101},
	'n': {0b000, 0b110, 0b101, 0b101, 0b101},
	'o': {0b000, 0b111, 0b101, 0b101, 0b111},
	'p': {0b000, 0b111, 0b101, 0b110, 0b100},
	'r': {0b000, 0b110, 0b101, 0b110, 0b101},
	's': {0b000, 0b111, 0b110, 0b011, 0b111},
	't': {0b010, 0b111, 0b010, 0b010, 0b011},
	'u': {0b000, 0b101, 0b101, 0b101, 0b111},
	'v': {0b000, 0b101, 0b101, 0b111, 0b010},
	'w': {0b000, 0b101, 0b111, 0b111, 0b101},
	'x': {0b000, 0b101, 0b010, 0b010, 0b101},
	'y': {0b000, 0b101, 0b111, 0b001, 0b111},
	'z': {0b000, 0b111, 0b011, 0b110, 0b111},
	' ': {0b000, 0b000, 0b000, 0b000, 0b000},
	'.': {0b000, 0b000, 0b000, 0b000, 0b010},
	'-': {0b000, 0b000, 0b111, 0b000, 0b000},
}

func drawSimpleText(img *image.RGBA, s string, x, y int, c color.RGBA) {
	cx := x
	scale := 2 // pixel scale factor
	for _, ch := range s {
		glyph, ok := font3x5[ch]
		if !ok {
			glyph = font3x5[' ']
		}
		for row := 0; row < 5; row++ {
			for col := 0; col < 3; col++ {
				if glyph[row]&(1<<uint(2-col)) != 0 {
					for dy := 0; dy < scale; dy++ {
						for dx := 0; dx < scale; dx++ {
							setPixel(img, cx+col*scale+dx, y+row*scale+dy, c)
						}
					}
				}
			}
		}
		cx += 4 * scale
	}
}

// drawYLabel draws a float64 value on the Y axis
func drawYLabel(img *image.RGBA, v float64, y int, c color.RGBA) {
	s := fmt.Sprintf("%.2f", v)
	drawSimpleText(img, s, 2, y-4, c)
}

// ─── GenerateChart ────────────────────────────────────────────────────────────
func GenerateChart(ticker string, bars []Bar, ind Indicators, label string) error {
	n := len(bars)
	if n == 0 {
		return fmt.Errorf("no data")
	}

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	draw.Draw(img, img.Bounds(), &image.Uniform{colBackground}, image.Point{}, draw.Src)

	price, macd, stoch, corr := makePanels()

	// ── PRICE PANEL ──────────────────────────────────────────────────────────
	// Compute price range (include Keltner bounds)
	priceMin, priceMax := bars[0].L, bars[0].H
	for i := 0; i < n; i++ {
		if bars[i].L < priceMin {
			priceMin = bars[i].L
		}
		if bars[i].H > priceMax {
			priceMax = bars[i].H
		}
		if ind.KeltUp[i] > priceMax {
			priceMax = ind.KeltUp[i]
		}
		if ind.KeltDn[i] < priceMin {
			priceMin = ind.KeltDn[i]
		}
	}
	pad := (priceMax - priceMin) * 0.06
	priceMin -= pad
	priceMax += pad

	drawPanelBG(img, price, 6, priceMin, priceMax)

	// Volume sub-pane occupies the bottom 15% of the price panel
	volPanelH := int(float64(price.y1-price.y0) * 0.15)
	volPanel := panel{price.y1 - volPanelH, price.y1}
	volMin := 0.0
	_, volMax := minMaxSlice(ind.Volume)

	// Ichimoku cloud — filled polygons column by column
	cloudOffset := 26
	for i := 0; i < n-1; i++ {
		ci := i + cloudOffset
		if ci >= n {
			break
		}
		x0 := scaleX(i, n)
		x1 := scaleX(i+1, n)
		a0 := scaleY(ind.SpanA[ci], priceMin, priceMax, price)
		b0 := scaleY(ind.SpanB[ci], priceMin, priceMax, price)
		isBull := ind.SpanA[ci] > ind.SpanB[ci]
		top := min2(a0, b0)
		bot := max2(a0, b0)
		col := colCloudBull
		if !isBull {
			col = colCloudBear
		}
		for x := x0; x < x1; x++ {
			for y := top; y <= bot; y++ {
				blendPixel(img, x, y, col)
			}
		}
	}

	// Keltner fill band
	for i := 0; i < n-1; i++ {
		x0 := scaleX(i, n)
		x1 := scaleX(i+1, n)
		top := min2(
			scaleY(ind.KeltUp[i], priceMin, priceMax, price),
			scaleY(ind.KeltUp[i+1], priceMin, priceMax, price),
		)
		bot := max2(
			scaleY(ind.KeltDn[i], priceMin, priceMax, price),
			scaleY(ind.KeltDn[i+1], priceMin, priceMax, price),
		)
		for x := x0; x < x1; x++ {
			for y := top; y <= bot; y++ {
				blendPixel(img, x, y, colKeltFill)
			}
		}
	}

	// Ichimoku Tenkan + Kijun lines (below candles)
	drawSeries(img, ind.Tenkan, priceMin, priceMax, price, colTenkan)
	drawSeries(img, ind.Kijun, priceMin, priceMax, price, colKijun)

	// Candles
	candleW := (imgW - marginL - marginR) / n
	if candleW < 1 {
		candleW = 1
	}
	for i := 0; i < n; i++ {
		x := scaleX(i, n)
		hi := scaleY(bars[i].H, priceMin, priceMax, price)
		lo := scaleY(bars[i].L, priceMin, priceMax, price)
		op := scaleY(bars[i].O, priceMin, priceMax, price)
		cl := scaleY(bars[i].C, priceMin, priceMax, price)
		bodyTop := min2(op, cl)
		bodyBot := max2(op, cl)
		c := colCandleUp
		if bars[i].C < bars[i].O {
			c = colCandleDn
		}
		// Wick
		vLine(img, x, hi, lo, c)
		// Body (fill width)
		hw := candleW / 2
		if hw < 1 {
			hw = 1
		}
		for bx := x - hw; bx <= x+hw; bx++ {
			vLine(img, bx, bodyTop, bodyBot, c)
		}
	}

	// Keltner channel lines on top
	drawSeries(img, ind.KeltUp, priceMin, priceMax, price, colKeltUp)
	drawSeries(img, ind.KeltMid, priceMin, priceMax, price, colKeltMid)
	drawSeries(img, ind.KeltDn, priceMin, priceMax, price, colKeltDn)

	// Volume bars at bottom of price panel
	for i := 0; i < n; i++ {
		x := scaleX(i, n)
		y0 := scaleY(ind.Volume[i], volMin, volMax, volPanel)
		y1 := volPanel.y1
		vc := colVolUp
		if bars[i].C < bars[i].O {
			vc = colVolDn
		}
		hw := candleW / 2
		if hw < 1 {
			hw = 1
		}
		for bx := x - hw; bx <= x+hw; bx++ {
			vLineBlend(img, bx, y0, y1, vc)
		}
	}

	// Panel label
	last := bars[n-1].C
	drawSimpleText(img, fmt.Sprintf("%s - %s  Last: %.2f", ticker, label, last), marginL+4, price.y0+4, colPanelLabel)

	// Y axis labels (price)
	for g := 0; g <= 4; g++ {
		v := priceMin + float64(g)*(priceMax-priceMin)/4
		y := scaleY(v, priceMin, priceMax, price)
		drawYLabel(img, v, y, colText)
		hLine(img, marginL-3, marginL, y, colAxis)
	}

	drawDateLabels(img, bars, price)

	// ── MACD PANEL ────────────────────────────────────────────────────────────
	macdMin, macdMax := minMaxSlice(ind.MACD)
	hMin, hMax := minMaxSlice(ind.Hist)
	if hMin < macdMin {
		macdMin = hMin
	}
	if hMax > macdMax {
		macdMax = hMax
	}
	sMin, sMax := minMaxSlice(ind.Signal)
	if sMin < macdMin {
		macdMin = sMin
	}
	if sMax > macdMax {
		macdMax = sMax
	}
	pad = math.Abs(macdMax-macdMin) * 0.1
	macdMin -= pad
	macdMax += pad

	drawPanelBG(img, macd, 4, macdMin, macdMax)

	// Zero line
	y0 := scaleY(0, macdMin, macdMax, macd)
	hLine(img, marginL, imgW-marginR, y0, colZero)

	// Histogram
	for i := 0; i < n; i++ {
		x := scaleX(i, n)
		yH := scaleY(ind.Hist[i], macdMin, macdMax, macd)
		yZ := scaleY(0, macdMin, macdMax, macd)
		hc := colHistPos
		if ind.Hist[i] < 0 {
			hc = colHistNeg
		}
		vLine(img, x, yH, yZ, hc)
	}

	drawSeries(img, ind.MACD, macdMin, macdMax, macd, colMACDLine)
	drawSeries(img, ind.Signal, macdMin, macdMax, macd, colSignalLine)
	drawSimpleText(img, "MACD 12.26.9", marginL+4, macd.y0+4, colPanelLabel)

	// ── STOCHASTIC PANEL ─────────────────────────────────────────────────────
	stochMin, stochMax := 0.0, 100.0
	drawPanelBG(img, stoch, 4, stochMin, stochMax)

	// Fill zones
	y80 := scaleY(80, stochMin, stochMax, stoch)
	y20 := scaleY(20, stochMin, stochMax, stoch)
	fillRect(img, marginL, stoch.y0, imgW-marginR, y80, colOBFill)
	fillRect(img, marginL, y20, imgW-marginR, stoch.y1, colOSFill)

	// OB/OS lines dashed
	for x := marginL; x < imgW-marginR; x += 6 {
		hLine(img, x, x+3, y80, colOB)
		hLine(img, x, x+3, y20, colOS)
	}

	drawSeries(img, ind.StochK, stochMin, stochMax, stoch, colStochK)
	drawSeries(img, ind.StochD, stochMin, stochMax, stoch, colStochD)
	drawSimpleText(img, "Stoch 80.20.10.3", marginL+4, stoch.y0+4, colPanelLabel)

	// ── CORRELATION PANEL ────────────────────────────────────────────────────
	corrMin, corrMax := minMaxSlice(ind.Correlation)
	padC := (corrMax - corrMin) * 0.15
	corrMin -= padC
	corrMax += padC
	drawPanelBG(img, corr, 2, corrMin, corrMax)

	// Zero line
	yZ2 := scaleY(0, corrMin, corrMax, corr)
	hLine(img, marginL, imgW-marginR, yZ2, colZero)
	drawSeries(img, ind.Correlation, corrMin, corrMax, corr, colCorr)
	drawSimpleText(img, "Corr SPY 10", marginL+4, corr.y0+4, colPanelLabel)

	_ = volMin // used above

	// ── SAVE ─────────────────────────────────────────────────────────────────
	os.MkdirAll("charts", os.ModePerm)

	// Sanitize label for filename
	safeLabel := ""
	for _, c := range label {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			safeLabel += string(c)
		} else {
			safeLabel += "_"
		}
	}

	path := fmt.Sprintf("charts/%s_%s_%d.png", ticker, safeLabel, time.Now().UnixNano())
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// latestChart returns the path of the most recently created chart file
// matching charts/<ticker>_<label>_*.png
func latestChart(ticker, label string) string {
	pattern := fmt.Sprintf("charts/%s_%s_*.png", ticker, label)
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return fmt.Sprintf("charts/%s_%s.png", ticker, label)
	}
	sort.Strings(matches)
	return matches[len(matches)-1]
}