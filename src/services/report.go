package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"framely/src/config"
	"framely/src/models"
)

// reportservice holds configuration and output directory for report operations
type ReportService struct {
	config    *config.Config
	outputDir string
}

// newreportservice creates a new reportservice instance with the given config
func NewReportService(cfg *config.Config) *ReportService {
	return &ReportService{
		config:    cfg,
		outputDir: cfg.OutputDir,
	}
}

// loadexistingreport loads the existing report from the output directory
func (rs *ReportService) LoadExistingReport() (*models.Report, error) {
	reportPath := filepath.Join(rs.outputDir, config.REPORT_FILE)

	data, err := os.ReadFile(reportPath)
	if err != nil {
		return nil, err
	}

	var report models.Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}

	return &report, nil
}

// generatereport generates a new report by combining existing and new results, calculates statistics, saves json and summary
func (rs *ReportService) GenerateReport(session *models.CrawlSession, existingReport *models.Report) error {
	var allResults []models.ScreenshotResult

	if existingReport != nil {
		allResults = append(allResults, existingReport.Results...)
	}

	sessionResults := session.GetResults()
	allResults = append(allResults, sessionResults...)

	successCount := 0
	failCount := 0
	totalDuration := int64(0)
	totalFileSize := int64(0)

	for _, result := range allResults {
		if result.Success {
			successCount++
			totalFileSize += result.FileSize
		}
		if !result.Success {
			failCount++
		}
		totalDuration += result.Duration
	}

	averagePageSize := int64(0)
	if successCount > 0 {
		averagePageSize = totalFileSize / int64(successCount)
	}

	timestamp := time.Now()
	var lastUpdate *time.Time
	if len(sessionResults) > 0 {
		lastUpdate = &timestamp
	}

	report := models.Report{
		BaseURL:               rs.config.BaseURL,
		TotalPages:            len(allResults),
		SuccessfulScreenshots: successCount,
		FailedScreenshots:     failCount,
		Timestamp:             timestamp,
		LastUpdate:            lastUpdate,
		NewPagesInThisRun:     len(sessionResults),
		TotalDuration:         totalDuration,
		AveragePageSize:       averagePageSize,
		Results:               allResults,
	}

	if err := rs.saveJSONReport(report); err != nil {
		return fmt.Errorf("failed to save JSON report: %w", err)
	}

	if err := rs.generateSummary(report, sessionResults); err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	return nil
}

// savejsonreport saves the report as json to the output directory
func (rs *ReportService) saveJSONReport(report models.Report) error {
	reportPath := filepath.Join(rs.outputDir, config.REPORT_FILE)

	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(reportPath, reportJSON, 0644); err != nil {
		return err
	}

	return nil
}

// generatesummary generates and saves the summary file with report details
func (rs *ReportService) generateSummary(report models.Report, newResults []models.ScreenshotResult) error {
	summaryPath := filepath.Join(rs.outputDir, config.SUMMARY_FILE)

	summary := rs.buildSummaryContent(report, newResults)

	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		return err
	}

	return nil
}

// buildsummarycontent builds the summary content as a formatted string with statistics and results
func (rs *ReportService) buildSummaryContent(report models.Report, newResults []models.ScreenshotResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\033[36m> Site: %s\n\033[0m", report.BaseURL))
	sb.WriteString(fmt.Sprintf("\033[36m> Generated: %s\n\033[0m", report.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("\033[36m> Total Pages: %d\n\033[0m", report.TotalPages))
	sb.WriteString(fmt.Sprintf("\033[32m> Successful: %d\n\033[0m", report.SuccessfulScreenshots))
	sb.WriteString(fmt.Sprintf("\033[31m> Failed: %d\n\033[0m", report.FailedScreenshots))
	sb.WriteString(fmt.Sprintf("\033[36m> New in this run: %d\n\033[0m", report.NewPagesInThisRun))
	sb.WriteString(fmt.Sprintf("\033[36m> Total Duration: %.2f seconds\n\033[0m", float64(report.TotalDuration)/1000.0))
	sb.WriteString(fmt.Sprintf("\033[36m> Average Page Size: %.2f KB\n\n\033[0m", float64(report.AveragePageSize)/1024.0))

	if len(newResults) > 0 {
		sb.WriteString(fmt.Sprintf("\033[36m> Newly added pages (%d):\n\033[0m", len(newResults)))
		for _, result := range newResults {
			if result.Success {
				details := fmt.Sprintf("%.2fKB, %dms", float64(result.FileSize)/1024, result.Duration)
				sb.WriteString(fmt.Sprintf("\033[32m> SUCCESS %s -> %s (%s)\n\033[0m", result.URL, result.Filename, details))
			}
			if !result.Success {
				sb.WriteString(fmt.Sprintf("\033[31m> FAILED %s -> %s (%s)\n\033[0m", result.URL, result.Filename, result.Error))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\033[36m> All successful pages:\n\033[0m")
	for _, result := range report.Results {
		if result.Success {
			sb.WriteString(fmt.Sprintf("\033[32m> SUCCESS %s -> %s (%.2fKB)\n\033[0m",
				result.URL,
				result.Filename,
				float64(result.FileSize)/1024,
			))
		}
	}

	if report.FailedScreenshots > 0 {
		sb.WriteString("\n\033[36m> Failed pages:\n\033[0m")
		for _, result := range report.Results {
			if !result.Success {
				sb.WriteString(fmt.Sprintf("\033[31m> FAILED %s - %s\n\033[0m", result.URL, result.Error))
			}
		}
	}

	return sb.String()
}

// getexistingurls returns a map of existing urls from the report
func (rs *ReportService) GetExistingURLs() (map[string]bool, error) {
	existingURLs := make(map[string]bool)

	report, err := rs.LoadExistingReport()
	if err != nil {
		return existingURLs, nil
	}

	for _, result := range report.Results {
		if result.Success {
			existingURLs[result.URL] = true
		}
	}

	return existingURLs, nil
}

// ensureoutputdirectory creates the output directory if it does not exist
func (rs *ReportService) EnsureOutputDirectory() error {
	if _, err := os.Stat(rs.outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(rs.outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}
	return nil
}
