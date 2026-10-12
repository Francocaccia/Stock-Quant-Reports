package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	dataCache = make(map[string]cachedBars)
	cacheMu   sync.RWMutex
	avMu      sync.Mutex
	lastAV    time.Time
)

type cachedBars struct {
	Bars      []Bar
	ExpiresAt time.Time
}

type Bar struct {
	T time.Time
	O float64
	H float64
	L float64
	C float64
	V float64
}

func FetchData(ticker string, skipStooq *bool) ([]Bar, error) {
	symbol := strings.ToUpper(ticker)

	// Check Cache
	cacheMu.RLock()
	c, ok := dataCache[symbol]
	cacheMu.RUnlock()
	if ok && time.Now().Before(c.ExpiresAt) {
		log.Printf("✅ Cache hit for %s", symbol)
		return c.Bars, nil
	}

	var errTiingo, errStooq, errYahoo, errTwelve, errAV error
	var bars []Bar
	var err error

	// 1. Try Tiingo (Highly reliable)
	if os.Getenv("TIINGO_API_KEY") != "" {
		bars, err = FetchDailyOHLCVTiingo(symbol)
		if err == nil && len(bars) > 0 {
			sortAndCache(symbol, bars)
			return bars, nil
		}
		errTiingo = err
	} else {
		errTiingo = fmt.Errorf("TIINGO_API_KEY not set")
	}

	// 2. Try TwelveData
	if os.Getenv("TWELVEDATA_API_KEY") != "" {
		bars, err = FetchDailyOHLCVTwelveData(symbol)
		if err == nil && len(bars) > 0 {
			sortAndCache(symbol, bars)
			return bars, nil
		}
		errTwelve = err
	} else {
		errTwelve = fmt.Errorf("TWELVEDATA_API_KEY not set")
	}

	// 3. Try Stooq (Free Scraping - trying multiple suffixes)
	if skipStooq == nil || !*skipStooq {
		altSymbols := []string{symbol}
		if !strings.Contains(symbol, ".") {
			altSymbols = []string{symbol + ".US", symbol}
		}
		for _, s := range altSymbols {
			bars, err = FetchDailyOHLCVStooq(s)
			if err == nil && len(bars) > 0 {
				sortAndCache(symbol, bars)
				return bars, nil
			}
			errStooq = err
		}
		if skipStooq != nil && strings.Contains(errStooq.Error(), "timeout") {
			*skipStooq = true
		}
	} else {
		errStooq = fmt.Errorf("skipped")
	}

	// 4. Try Yahoo Finance
	bars, err = FetchDailyOHLCVYahoo(symbol)
	if err == nil && len(bars) > 0 {
		sortAndCache(symbol, bars)
		return bars, nil
	}
	errYahoo = err

	// 5. Try AlphaVantage (Smart Retry)
	if os.Getenv("ALPHAVANTAGE_API_KEY") != "" {
		bars, err = FetchDailyOHLCVAlphaVantage(symbol)
		if err != nil && (strings.Contains(err.Error(), "rate limit") || strings.Contains(err.Error(), "status 429")) {
			log.Printf("⏳ AlphaVantage rate limit hit for %s. Waiting 5s for retry...", symbol)
			time.Sleep(5 * time.Second)
			bars, err = FetchDailyOHLCVAlphaVantage(symbol)
		}
		if err == nil && len(bars) > 0 {
			sortAndCache(symbol, bars)
			return bars, nil
		}
		errAV = err
	} else {
		errAV = fmt.Errorf("ALPHAVANTAGE_API_KEY not set")
	}

	return nil, fmt.Errorf("all sources failed for %s.\nTiingo: %v\nTwelveData: %v\nStooq: %v\nYahoo: %v\nAV: %v", ticker, errTiingo, errTwelve, errStooq, errYahoo, errAV)
}

func sortAndCache(ticker string, bars []Bar) {
	sortBars(bars)
	saveToCache(ticker, bars)
}

func sortBars(bars []Bar) {
	sort.Slice(bars, func(i, j int) bool {
		return bars[i].T.Before(bars[j].T)
	})
}

func saveToCache(ticker string, bars []Bar) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	dataCache[ticker] = cachedBars{
		Bars:      bars,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

func FetchDailyOHLCVStooq(symbol string) ([]Bar, error) {
	url := "https://stooq.com/q/d/l/?s=" + symbol + "&i=d"
	log.Printf("📥 Fetching from Stooq: %s", url)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://stooq.com/q/d/?s="+symbol)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stooq connection error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stooq status %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)
	_, err = reader.Read() // Skip header
	if err != nil {
		return nil, fmt.Errorf("stooq: error reading header")
	}

	var bars []Bar
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 6 {
			continue
		}

		t, _ := time.Parse("2006-01-02", record[0])
		o, _ := strconv.ParseFloat(record[1], 64)
		h, _ := strconv.ParseFloat(record[2], 64)
		l, _ := strconv.ParseFloat(record[3], 64)
		c, _ := strconv.ParseFloat(record[4], 64)
		v, _ := strconv.ParseFloat(record[5], 64)

		bars = append(bars, Bar{T: t, O: o, H: h, L: l, C: c, V: v})
	}

	if len(bars) == 0 {
		return nil, fmt.Errorf("stooq: empty response")
	}

	// Stooq is usually newest-first
	for i, j := 0, len(bars)-1; i < j; i, j = i+1, j-1 {
		bars[i], bars[j] = bars[j], bars[i]
	}

	return bars, nil
}

func FetchDailyOHLCVAlphaVantage(ticker string) ([]Bar, error) {
	apiKey := os.Getenv("ALPHAVANTAGE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ALPHAVANTAGE_API_KEY not set")
	}

	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=%s&outputsize=full&apikey=%s&datatype=csv", ticker, apiKey)
	log.Printf("📥 Fetching from AlphaVantage: %s", strings.Replace(url, apiKey, "REDACTED", 1))

	avMu.Lock()
	elapsed := time.Since(lastAV)
	if elapsed < 12*time.Second {
		wait := 12*time.Second - elapsed
		log.Printf("⏳ AlphaVantage Stagger: waiting %v for %s...", wait, ticker)
		time.Sleep(wait)
	}
	lastAV = time.Now()
	avMu.Unlock()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if strings.Contains(bodyStr, "Error Message") {
		return nil, fmt.Errorf("alphavantage: invalid symbol")
	}
	if strings.Contains(bodyStr, "Note") || strings.Contains(bodyStr, "Information") {
		return nil, fmt.Errorf("alphavantage rate limit")
	}

	reader := csv.NewReader(strings.NewReader(bodyStr))
	_, _ = reader.Read() // Skip header

	var bars []Bar
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 6 {
			continue
		}

		t, _ := time.Parse("2006-01-02", record[0])
		o, _ := strconv.ParseFloat(record[1], 64)
		h, _ := strconv.ParseFloat(record[2], 64)
		l, _ := strconv.ParseFloat(record[3], 64)
		c, _ := strconv.ParseFloat(record[4], 64)
		v, _ := strconv.ParseFloat(record[5], 64)

		bars = append(bars, Bar{T: t, O: o, H: h, L: l, C: c, V: v})
	}

	return bars, nil
}

func FetchDailyOHLCVYahoo(symbol string) ([]Bar, error) {
	symbol = strings.ReplaceAll(symbol, ".US", "")
	now := time.Now().Unix()
	tenYearsAgo := time.Now().AddDate(-10, 0, 0).Unix()

	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/download/%s?period1=%d&period2=%d&interval=1d&events=history&includeAdjustedClose=true", symbol, tenYearsAgo, now)
	log.Printf("📥 Fetching from Yahoo: %s", url)

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/csv,text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Referer", "https://finance.yahoo.com/quote/"+symbol+"/history")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yahoo status %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)
	_, _ = reader.Read() // Skip header

	var bars []Bar
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 7 {
			continue
		}

		t, _ := time.Parse("2006-01-02", record[0])
		o, _ := strconv.ParseFloat(record[1], 64)
		h, _ := strconv.ParseFloat(record[2], 64)
		l, _ := strconv.ParseFloat(record[3], 64)
		c, _ := strconv.ParseFloat(record[5], 64) // Adj Close
		v, _ := strconv.ParseFloat(record[6], 64)

		bars = append(bars, Bar{T: t, O: o, H: h, L: l, C: c, V: v})
	}

	return bars, nil
}

func FetchDailyOHLCVTiingo(symbol string) ([]Bar, error) {
	apiKey := os.Getenv("TIINGO_API_KEY")
	symbol = strings.ReplaceAll(symbol, ".US", "")
	// Fetch last 10 years by default for comprehensive charts
	startDate := time.Now().AddDate(-10, 0, 0).Format("2006-01-02")
	url := fmt.Sprintf("https://api.tiingo.com/tiingo/daily/%s/prices?token=%s&startDate=%s&format=csv", symbol, apiKey, startDate)
	log.Printf("📥 Fetching from Tiingo: %s", symbol)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tiingo status %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)
	_, _ = reader.Read() // Skip header

	var bars []Bar
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 6 {
			continue
		}

		t, _ := time.Parse("2006-01-02", record[0])
		c, _ := strconv.ParseFloat(record[1], 64)
		h, _ := strconv.ParseFloat(record[2], 64)
		l, _ := strconv.ParseFloat(record[3], 64)
		o, _ := strconv.ParseFloat(record[4], 64)
		v, _ := strconv.ParseFloat(record[5], 64)

		bars = append(bars, Bar{T: t, O: o, H: h, L: l, C: c, V: v})
	}

	return bars, nil
}

func FetchDailyOHLCVTwelveData(ticker string) ([]Bar, error) {
	apiKey := os.Getenv("TWELVEDATA_API_KEY")
	symbol := strings.ToUpper(ticker)
	symbol = strings.ReplaceAll(symbol, ".US", "")

	url := fmt.Sprintf("https://api.twelvedata.com/time_series?symbol=%s&interval=1day&outputsize=5000&apikey=%s&format=csv", symbol, apiKey)
	log.Printf("📥 Fetching from TwelveData: %s", symbol)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("twelvedata status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if strings.Contains(bodyStr, "error") {
		return nil, fmt.Errorf("twelvedata error")
	}

	lines := strings.Split(bodyStr, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("twelvedata: no data")
	}

	separator := ";"
	if strings.Contains(lines[0], ",") {
		separator = ","
	}

	var bars []Bar
	for i := 1; i < len(lines); i++ {
		record := strings.Split(strings.TrimSpace(lines[i]), separator)
		if len(record) < 5 {
			continue
		}

		t, err := time.Parse("2006-01-02", record[0])
		if err != nil {
			continue
		}
		o, _ := strconv.ParseFloat(record[1], 64)
		h, _ := strconv.ParseFloat(record[2], 64)
		l, _ := strconv.ParseFloat(record[3], 64)
		c, _ := strconv.ParseFloat(record[4], 64)
		v := 0.0
		if len(record) > 5 {
			v, _ = strconv.ParseFloat(record[5], 64)
		}

		bars = append(bars, Bar{T: t, O: o, H: h, L: l, C: c, V: v})
	}

	return bars, nil
}

func FilterLastYears(data []Bar, years int) []Bar {
	if len(data) == 0 {
		return data
	}
	cutoff := data[len(data)-1].T.AddDate(-years, 0, 0)
	var filtered []Bar
	for _, b := range data {
		if b.T.After(cutoff) {
			filtered = append(filtered, b)
		}
	}
	return filtered
}