package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var db *MongoClient

type GenerateRequest struct {
	Ticker    string `json:"ticker"`
	Benchmark string `json:"benchmark"`
	Email     string `json:"email"`
	Analyst   string `json:"analyst"` // display name from Google
}

type GenerateResponse struct {
	PDFPath string `json:"pdfPath"`
	Signal  string `json:"signal"`
	Error   string `json:"error,omitempty"`
}

type BatchResult struct {
	Ticker  string `json:"ticker"`
	PDFPath string `json:"pdfPath,omitempty"`
	Signal  string `json:"signal,omitempty"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

type BatchResponse struct {
	Results []BatchResult `json:"results"`
	Error   string        `json:"error,omitempty"`
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Ensure output directories exist for Cloud Run environment
	os.MkdirAll("./reports", os.ModePerm)
	os.MkdirAll("./charts", os.ModePerm)

	// Init MongoDB if available
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI != "" {
		var err error
		db, err = NewMongoClient(mongoURI)
		if err != nil {
			log.Printf("⚠️ Failed to connect to MongoDB: %v", err)
		} else {
			fmt.Println("✅ Connected to MongoDB Atlas")
		}
	} else {
		log.Println("⚠️ MONGODB_URI not set, running without database")
	}

	// Root handler for Render Health Check (avoids deployment failure)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(&w)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Quant API is Live 🚀")
	})

	// Serve the API endpoints
	http.HandleFunc("/api/generate", handleGenerate)
	http.HandleFunc("/api/batch", handleBatch)
	http.HandleFunc("/api/history", handleHistory)
	http.HandleFunc("/api/admin/history", handleAdminHistory)

	// Serve the generated PDFs WITH CORS (Attempt Mongo DB first, fallback to disk)
	http.HandleFunc("/reports/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(&w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		reqPath := r.URL.Path // e.g. "/reports/AAPL_report_date.pdf"

		if db != nil {
			data, err := db.GetPDFData(reqPath)
			if err == nil && len(data) > 0 {
				w.Header().Set("Content-Type", "application/pdf")
				w.Write(data)
				return
			}
		}

		// Fallback to local file
		http.ServeFile(w, r, "."+reqPath)
	})

	fmt.Printf("🚀 Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req GenerateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ticker := strings.ToUpper(req.Ticker)
	benchmark := req.Benchmark
	if benchmark == "" {
		benchmark = "SPY"
	}
	analyst := req.Analyst
	if analyst == "" {
		analyst = req.Email
	}
	if analyst == "" {
		analyst = "Unknown"
	}

	pdfPath, signal, err := processSingleTicker(ticker, benchmark, analyst, req.Email)
	if err != nil {
		json.NewEncoder(w).Encode(GenerateResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GenerateResponse{
		PDFPath: pdfPath,
		Signal:  signal,
	})
}

func handleBatch(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	benchmark := r.FormValue("benchmark")
	if benchmark == "" {
		benchmark = "SPY"
	}
	analyst := r.FormValue("analyst")
	if analyst == "" {
		analyst = r.FormValue("email")
	}
	if analyst == "" {
		analyst = "Unknown"
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		json.NewEncoder(w).Encode(BatchResponse{Error: "Failed to read uploaded file"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		json.NewEncoder(w).Encode(BatchResponse{Error: "Failed to read file content"})
		return
	}

	lines := strings.Split(string(content), "\n")
	var results []BatchResult

	for _, line := range lines {
		ticker := strings.TrimSpace(line)
		ticker = strings.ReplaceAll(ticker, ",", "")
		ticker = strings.ReplaceAll(ticker, "\r", "") // CRITICAL for Windows created files
		if ticker == "" {
			continue
		}

		email := r.FormValue("email")

		pdfPath, signal, err := processSingleTicker(ticker, benchmark, analyst, email)
		if err != nil {
			results = append(results, BatchResult{Ticker: ticker, Status: "Failed", Error: err.Error()})
		} else {
			results = append(results, BatchResult{Ticker: ticker, Status: "Success", PDFPath: pdfPath, Signal: signal})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BatchResponse{Results: results})
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if db == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	email := r.URL.Query().Get("email")
	ticker := r.URL.Query().Get("ticker")
	signal := r.URL.Query().Get("signal")
	dateFrom := r.URL.Query().Get("from")
	dateTo := r.URL.Query().Get("to")
	signals, err := db.GetSignals(email, ticker, signal, dateFrom, dateTo, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if signals == nil {
		signals = []SignalDoc{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signals)
}

func handleAdminHistory(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if db == nil {
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	// Fetch ALL history with optional filters
	ticker := r.URL.Query().Get("ticker")
	signal := r.URL.Query().Get("signal")
	dateFrom := r.URL.Query().Get("from")
	dateTo := r.URL.Query().Get("to")
	signals, err := db.GetSignals("", ticker, signal, dateFrom, dateTo, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if signals == nil {
		signals = []SignalDoc{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signals)
}

func getNYTime() time.Time {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return time.Now()
	}
	return time.Now().In(loc)
}

func processSingleTicker(ticker, benchmark, analyst, email string) (string, string, error) {
	var skipStooq bool

	type fetchResult struct {
		bars []Bar
		err  error
	}

	tickerChan := make(chan fetchResult)
	benchChan := make(chan fetchResult)

	// Fetch Ticker and Benchmark in PARALLEL
	go func() {
		b, e := FetchData(ticker, &skipStooq)
		tickerChan <- fetchResult{b, e}
	}()

	go func() {
		b, e := FetchData(benchmark, &skipStooq)
		benchChan <- fetchResult{b, e}
	}()

	tRes := <-tickerChan
	if tRes.err != nil {
		return "", "", tRes.err
	}
	daily := tRes.bars

	bRes := <-benchChan
	if bRes.err != nil {
		return "", "", bRes.err
	}
	spyDaily := bRes.bars

	// 1. Resample full data
	weekly := ResampleToWeekly(daily)
	monthly := ResampleToMonthly(daily)

	spyWeekly := ResampleToWeekly(spyDaily)
	spyMonthly := ResampleToMonthly(spyDaily)

	// 2. Compute indicators on FULL data (avoids cold-start)
	wInd := ComputeAllIndicators(weekly, spyWeekly)
	mInd := ComputeAllIndicators(monthly, spyMonthly)

	// 3. Filter for report (FUND STANDARD: last 3 years)
	weeklyFiltered := FilterLastYears(weekly, 3)
	monthlyFiltered := FilterLastYears(monthly, 3)

	// Keep indicators aligned with filtered data
	wIndFiltered := FilterIndicators(wInd, len(weekly), len(weeklyFiltered))
	mIndFiltered := FilterIndicators(mInd, len(monthly), len(monthlyFiltered))

	weeklyState := BuildState(weeklyFiltered, wIndFiltered)
	monthlyState := BuildState(monthlyFiltered, mIndFiltered)

	weeklyText := WeeklyNarrative(weeklyState, wIndFiltered, weeklyFiltered)
	monthlyText := MonthlyNarrative(monthlyState, mIndFiltered, monthlyFiltered, weeklyState)

	signal, rationale := GenerateSignal(weeklyState, monthlyState)

	// Generate charts — names are timestamped, grab the newest file
	GenerateChart(ticker, weeklyFiltered, wIndFiltered, "weekly")
	GenerateChart(ticker, monthlyFiltered, mIndFiltered, "monthly")

	weeklyChart := latestChart(ticker, "weekly")
	monthlyChart := latestChart(ticker, "monthly")

	outputPath, err := GeneratePDF(
		ticker,
		analyst,
		weeklyText,
		monthlyText,
		signal,
		rationale,
		weeklyChart,
		monthlyChart,
	)

	if err != nil {
		return "", "", fmt.Errorf("Error generating PDF")
	}

	// Create relative path for frontend access
	relativePath := strings.Replace(outputPath, "\\", "/", -1)
	parts := strings.Split(relativePath, "reports/")
	downloadUrl := "/reports/" + parts[len(parts)-1]

	pdfData, _ := os.ReadFile(outputPath)

	// Save to MongoDB if available
	if db != nil {
		err := db.InsertSignal(SignalDoc{
			TS:      getNYTime(),
			Ticker:  ticker,
			Signal:  signal,
			Analyst: analyst,
			Email:   email,
			PDFFile: downloadUrl,
			PDFData: pdfData,
			Source:  "Multi-Provider",
			States: map[string]any{
				"weekly":  weeklyState,
				"monthly": monthlyState,
			},
		})
		if err != nil {
			log.Printf("⚠️ Failed to save report to DB: %v", err)
		}
	}

	return downloadUrl, signal, nil
}
