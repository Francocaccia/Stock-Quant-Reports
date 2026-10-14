package main

import (
	"os"
	"time"

	"github.com/wcharczuk/go-chart/v2"
)

func RenderSimpleChartPNG(bars []Bar, ind Indicators, title string, outPath string) error {
	n := len(bars)
	if n < 10 {
		return nil
	}

	x := make([]time.Time, n)
	close := make([]float64, n)
	tenkan := make([]float64, n)
	kijun := make([]float64, n)

	for i := 0; i < n; i++ {
		x[i] = bars[i].T
		close[i] = bars[i].C
		tenkan[i] = ind.Tenkan[i]
		kijun[i] = ind.Kijun[i]
	}

	graph := chart.Chart{
		Title: title,
		Background: chart.Style{
			Padding: chart.Box{Top: 30, Left: 10, Right: 10, Bottom: 10},
		},
		XAxis: chart.XAxis{
			ValueFormatter: chart.TimeDateValueFormatter,
		},
		YAxis: chart.YAxis{
			ValueFormatter: chart.FloatValueFormatter,
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name:    "Close",
				XValues: x,
				YValues: close,
			},
			chart.TimeSeries{
				Name:    "Tenkan",
				XValues: x,
				YValues: tenkan,
			},
			chart.TimeSeries{
				Name:    "Kijun",
				XValues: x,
				YValues: kijun,
			},
		},
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return graph.Render(chart.PNG, f)
}
