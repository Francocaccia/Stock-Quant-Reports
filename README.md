# Stock-Quant-Reports

A quantitative equity report generator built for a university student-managed investment fund (SMIF). Given a ticker (or a batch file of tickers), it pulls daily price history, computes technical indicators on weekly and monthly timeframes, and produces a two-page PDF with charts, a plain-English narrative and a rule-based signal.

This is an educational / portfolio project. **It is not investment advice.**

## Architecture

```
React + Vite  --(Firebase Auth / Hosting)-->  Go API  -->  third-party market-data APIs
   (web/)                                    (*.go)         (Tiingo, TwelveData, Stooq,
                                                             Yahoo Finance, Alpha Vantage)
                                                  |
                                                  v
                                               MongoDB
                                  (signal history + generated PDFs)
```

- **Frontend (`web/`)**: React + Vite single-page app. Sign-in with Google through Firebase Auth; hosted on Firebase Hosting.
- **Backend (repo root)**: a single Go HTTP server that fetches data, computes indicators, draws charts, writes PDFs and stores results.
- **Database**: MongoDB (optional; without it the API still generates reports but keeps no history).

## Features

- Single-ticker report and batch reports from an uploaded `.txt` / `.csv` list of tickers
- Benchmark-relative analysis (default benchmark: SPY)
- Market-data provider fallback chain, so one provider failing or rate-limiting does not block a report
- Per-user report history with filters and CSV export
- Optional admin view (activity per analyst, all reports) shown only to emails listed in `VITE_ADMIN_EMAILS`
- Signal and state history stored in MongoDB; PDFs are stored as binary so they survive ephemeral disks

## What a generated report contains

Each report is two A4 pages:

1. **Weekly chart (short-term view)**: a weekly chart with trend, momentum and volatility overlays, followed by a written interpretation of the weekly state.
2. **Monthly chart (long-term view)**: the monthly equivalent, followed by a written interpretation that takes the weekly state into account, and a final **Signal** line (a label plus a one-sentence rationale).

Indicators used include Ichimoku components, moving averages, MACD, Stochastic, ATR-based volatility, volume, and rolling correlation with the benchmark. The signal is produced by a deterministic rule set that combines the weekly and monthly states.

No sample reports, charts or provider data are included in this repository: third-party data licenses may not allow redistribution.

## Run locally

Requirements: Go 1.23+, Node 20+, and optionally MongoDB (local or Atlas).

### Environment variables (backend)

| Variable | Required | Description |
| --- | --- | --- |
| `PORT` | no | HTTP port (default `8080`) |
| `MONGODB_URI` | no | MongoDB connection string. If unset, history is disabled |
| `TIINGO_API_KEY` | no | Tiingo API key |
| `TWELVEDATA_API_KEY` | no | Twelve Data API key |
| `ALPHAVANTAGE_API_KEY` | no | Alpha Vantage API key |

Providers without a key are skipped; the keyless ones (Stooq, Yahoo Finance) may be rate-limited or blocked. See `.env.example`. The server reads the process environment directly (it does not load `.env` files), so export the variables in your shell.

### Get the code

```bash
git clone https://github.com/Francocaccia/Stock-Quant-Reports.git
cd Stock-Quant-Reports
```

### Backend

```bash
export MONGODB_URI=mongodb://localhost:27017   # optional
export TIINGO_API_KEY=...                      # optional, any provider key
go run .
```

### Frontend

```bash
cd web
cp .env.example .env.local     # set VITE_API_URL / VITE_ADMIN_EMAILS if needed
npm ci
npm run dev
```

| Variable | Description |
| --- | --- |
| `VITE_API_URL` | Base URL of the Go API (defaults to `http://localhost:8080`) |
| `VITE_ADMIN_EMAILS` | Comma-separated emails that see the admin UI (empty = nobody) |

The Firebase web config in `web/src/firebase.js` is public by design. To use your own Firebase project, replace it, update `web/.firebaserc`, and set the hosting `site` in `web/firebase.json`.

### Docker

```bash
docker build -t stock-quant-reports .
docker run -p 8080:8080 -e MONGODB_URI=... stock-quant-reports
```

## Security model and limitations

**This version has no server-side authentication.** Access control is enforced only in the frontend through Firebase Auth (a `@lynchburg.edu` domain check, plus the admin email list). Anyone who can reach the API can call every endpoint, including `/api/admin/history` and the history endpoint, which returns data for any email passed as a query parameter.

In addition:

- CORS is open (`Access-Control-Allow-Origin: *`).
- There is no rate limiting.
- Ticker and filter inputs are only lightly validated.

**Do not expose this API publicly** without first adding Firebase ID-token verification on the server, rate limits and a CORS allow-list.

## Roadmap

1. **Server-side authentication**: verify Firebase ID tokens in the Go API, enforce per-user and admin authorization on the server, add rate limits and a CORS allow-list.
2. Move the admin check and history filtering from client-supplied emails to token claims.
3. Tests for indicator math and signal rules.
4. Configurable indicator parameters and additional timeframes.
5. Structured logging and metrics.

## Known limitations

- Free data sources (Stooq, Yahoo Finance) are unofficial and can change or block without notice.
- Narratives and signals are rule-based heuristics, not a validated trading model; no backtesting is included.
- Charts are drawn with a minimal built-in bitmap font and simple rendering.
- Reports are written to local disk and to MongoDB; there is no cleanup or retention policy.
- Market data from third-party providers is subject to their own terms of use.

## License

MIT. See [LICENSE](LICENSE).
