package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf/v2"
)

// imgHeight returns the PDF height (mm) for an image whose native size is
// nativeW x nativeH points, when rendered at pdfW mm wide.
// Our chart is generated at 14" × 8" → ratio 8/14
func chartPDFHeight(pdfW float64) float64 {
	return pdfW * (8.0 / 14.0)
}

func GeneratePDF(
	ticker string,
	analyst string,
	weeklyText string,
	monthlyText string,
	signal string,
	rationale string,
	weeklyImgPath string,
	monthlyImgPath string,
) (string, error) {

	if err := os.MkdirAll("reports", os.ModePerm); err != nil {
		return "", err
	}

	loc_ny, _ := time.LoadLocation("America/New_York")
	reportTime := time.Now()
	if loc_ny != nil {
		reportTime = time.Now().In(loc_ny)
	}
	fileTime := reportTime.Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("reports/%s_report_%s.pdf", ticker, fileTime)

	// Use the analyst name as-is from Google; capitalise first letter of each word
	analystDisplay := toTitleCase(analyst)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(25, 20, 25) // Left 25 | Top 20 | Right 25
	pdf.SetAutoPageBreak(true, 15)

	// Usable width between margins
	usableW := 210.0 - 25.0 - 25.0 // A4 width 210mm − margins

	// ─────────────────────────────────────────────────────────────────
	// PAGE 1 — WEEKLY
	// ─────────────────────────────────────────────────────────────────
	pdf.AddPage()

	// --- Header: "Quant Team" left | analyst right (Arial 11 regular)
	pdf.SetFont("Arial", "", 11)
	pageW := 210.0
	marginL := 25.0
	marginR := 25.0
	pdf.SetX(marginL)
	pdf.Cell(0, 6, "Quant Team")
	pdf.SetXY(pageW-marginR-pdf.GetStringWidth(analystDisplay), pdf.GetY())
	pdf.Cell(pdf.GetStringWidth(analystDisplay), 6, analystDisplay)
	pdf.Ln(10)

	// --- Title block: "Quant Report" bold centered then ticker bold centered
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(usableW, 7, "Quant Report", "", 1, "C", false, 0, "")
	pdf.CellFormat(usableW, 7, ticker, "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// --- Section heading
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(usableW, 6, "Weekly Chart (Short-Term View):", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// --- Weekly chart image
	if weeklyImgPath != "" {
		if _, err := os.Stat(weeklyImgPath); err == nil {
			imgH := chartPDFHeight(usableW) // ~91 mm
			imgY := pdf.GetY()
			pdf.ImageOptions(weeklyImgPath, marginL, imgY, usableW, imgH, false,
				gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
			pdf.SetY(imgY + imgH + 5) // advance cursor below image
		}
	}

	// --- Weekly narrative: Arial 11, justified
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(usableW, 6, sanitizeForPDF(weeklyText), "", "J", false)

	// ─────────────────────────────────────────────────────────────────
	// PAGE 2 — MONTHLY + SIGNAL
	// ─────────────────────────────────────────────────────────────────
	pdf.AddPage()

	// --- Header
	pdf.SetFont("Arial", "", 11)
	pdf.SetX(marginL)
	pdf.Cell(0, 6, "Quant Team")
	pdf.SetXY(pageW-marginR-pdf.GetStringWidth(analystDisplay), pdf.GetY())
	pdf.Cell(pdf.GetStringWidth(analystDisplay), 6, analystDisplay)
	pdf.Ln(10)

	// --- Section heading
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(usableW, 6, "Monthly Chart (Long-Term View):", "", 1, "L", false, 0, "")
	pdf.Ln(2)

	// --- Monthly chart image
	if monthlyImgPath != "" {
		if _, err := os.Stat(monthlyImgPath); err == nil {
			imgH := chartPDFHeight(usableW)
			imgY := pdf.GetY()
			pdf.ImageOptions(monthlyImgPath, marginL, imgY, usableW, imgH, false,
				gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
			pdf.SetY(imgY + imgH + 5)
		}
	}

	// --- Monthly narrative
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(usableW, 6, sanitizeForPDF(monthlyText), "", "J", false)
	pdf.Ln(5)

	// --- Signal line: "Signal: " bold + signal + " - " + rationale regular
	labelW := pdf.GetStringWidth("Signal: ")
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(labelW+2, 6, "Signal: ", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(usableW-labelW-2, 6,
		sanitizeForPDF(fmt.Sprintf("%s - %s", strings.ToUpper(signal), rationale)), "", "J", false)

	// Save
	if err := pdf.OutputFileAndClose(fileName); err != nil {
		return "", err
	}
	absPath, _ := filepath.Abs(fileName)
	return absPath, nil
}

// toTitleCase converts "franco caccia", "Francocaccia", or email-derived strings
// into "Franco Caccia". Simple: split on space and capitalise each word.
func toTitleCase(s string) string {
	if s == "" {
		return "Unknown"
	}
	words := strings.Fields(s) // splits on any whitespace
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// sanitizeForPDF replaces UTF-8 characters that gofpdf cannot render
// (gofpdf uses Latin-1 encoding by default).
func sanitizeForPDF(s string) string {
	replacements := map[string]string{
		"\u2014": "-",   // em-dash
		"\u2013": "-",   // en-dash
		"\u2018": "'",  // left single quote
		"\u2019": "'",  // right single quote
		"\u201C": "\"", // left double quote
		"\u201D": "\"", // right double quote
		"\u2026": "...", // ellipsis
		"\u00e2\u0080\u0094": "-", // UTF-8 mis-encoded em-dash
	}
	for unicode, ascii := range replacements {
		s = strings.ReplaceAll(s, unicode, ascii)
	}
	return s
}
