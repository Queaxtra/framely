package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"framely/src/config"
	"framely/src/models"
	"framely/src/utils"
)

// appservice holds config and services for the main application logic
type AppService struct {
	config           *config.Config
	browserService   *BrowserService
	discoveryService *DiscoveryService
	reportService    *ReportService
	session          *models.CrawlSession
}

// newappservice creates a new appservice instance with initialized services
func NewAppService(cfg *config.Config) *AppService {
	return &AppService{
		config:           cfg,
		browserService:   NewBrowserService(cfg),
		discoveryService: NewDiscoveryService(cfg.BaseURL),
		reportService:    NewReportService(cfg),
		session:          models.NewCrawlSession(cfg.BaseURL),
	}
}

// initialize sets up the service, ensures output directory, tests connection, loads existing urls
func (as *AppService) Initialize() error {
	log.Printf("\033[36m> Initializing Framely Screenshot Service\033[0m")
	log.Printf("\033[36m> Target: %s\033[0m", as.config.BaseURL)
	log.Printf("\033[36m> Max Depth: %d\033[0m", as.config.MaxDepth)
	log.Printf("\033[36m> Parallel Workers: %d\033[0m", as.config.ParallelWorkers)

	if err := as.reportService.EnsureOutputDirectory(); err != nil {
		return fmt.Errorf("output directory creation failed: %w", err)
	}

	if err := as.browserService.TestConnection(as.config.BaseURL); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	existingURLs, err := as.reportService.GetExistingURLs()
	if err != nil {
		log.Printf("\033[31m> Could not load existing URLs: %s\033[0m", err.Error())
		existingURLs = make(map[string]bool)
	}
	for url := range existingURLs {
		as.session.MarkExisting(url)
	}
	log.Printf("\033[32m> Loaded %d existing URLs\033[0m", len(existingURLs))

	return nil
}

// discoverurls discovers urls from sitemap and robots.txt if enabled, adds them to session
func (as *AppService) DiscoverURLs() error {
	if !as.config.CheckSitemap && !as.config.CheckRobots {
		log.Printf("\033[36m> URL discovery disabled\033[0m")
		return nil
	}

	log.Printf("\033[36m> Starting URL discovery...\033[0m")

	discoveredURLs := as.discoveryService.DiscoverURLs(as.config.CheckSitemap, as.config.CheckRobots)

	for _, url := range discoveredURLs {
		normalizedURL := utils.NormalizeURL(url)
		if !as.session.IsVisited(normalizedURL) && !as.session.IsExisting(normalizedURL) {
			as.session.AddURL(url, 1)
		}
	}

	log.Printf("\033[32m> Discovery complete: %d URLs added to queue\033[0m", len(discoveredURLs))
	return nil
}

// crawlwebsite starts the crawl process, chooses between sequential or parallel based on config
func (as *AppService) CrawlWebsite() error {
	log.Printf("\033[36m> Starting website crawl...\033[0m")

	as.session.AddURL(as.config.BaseURL, 0)

	if as.config.ParallelWorkers > 1 {
		return as.runParallelCrawl()
	}

	return as.runSequentialCrawl()
}

// runsequentialcrawl processes urls one by one, captures screenshots, extracts links
func (as *AppService) runSequentialCrawl() error {
	log.Printf("\033[36m> Running sequential crawl...\033[0m")

	for {
		url, depth, hasNext := as.session.GetNextURL()
		if !hasNext {
			break
		}

		if depth > as.config.MaxDepth {
			continue
		}

		normalizedURL := utils.NormalizeURL(url)

		if as.session.IsVisited(normalizedURL) {
			continue
		}

		if as.session.IsExisting(normalizedURL) {
			log.Printf("\033[36m> Skipping existing: %s\033[0m", url)
			as.session.MarkVisited(normalizedURL)
			continue
		}

		if utils.ShouldSkipURL(url, as.config.SkipPatterns) {
			log.Printf("\033[36m> Skipping pattern match: %s\033[0m", url)
			as.session.MarkVisited(normalizedURL)
			continue
		}

		as.session.MarkVisited(normalizedURL)

		result := as.browserService.CaptureScreenshot(url)
		as.session.AddResult(result)

		if result.Success && depth < as.config.MaxDepth {
			links, err := as.browserService.ExtractLinks(url)
			if err != nil {
				log.Printf("\033[31m> Link extraction failed for %s: %s\033[0m", url, err.Error())
			}
			if err == nil {
				as.addNewLinksToQueue(links, depth+1)
			}
		}

		time.Sleep(time.Duration(as.config.RequestDelay) * time.Second)
	}

	return nil
}

// runparallelcrawl processes urls in parallel using workers, captures screenshots, extracts links
func (as *AppService) runParallelCrawl() error {
	log.Printf("\033[36m> Running parallel crawl with %d workers...\033[0m", as.config.ParallelWorkers)

	semaphore := make(chan struct{}, as.config.ParallelWorkers)
	var wg sync.WaitGroup
	resultsChan := make(chan models.ScreenshotResult, 100)

	go func() {
		for result := range resultsChan {
			as.session.AddResult(result)
		}
	}()

	for {
		url, depth, hasNext := as.session.GetNextURL()
		if !hasNext {
			break
		}

		if depth > as.config.MaxDepth {
			continue
		}

		normalizedURL := utils.NormalizeURL(url)

		if as.session.IsVisited(normalizedURL) || as.session.IsExisting(normalizedURL) {
			continue
		}

		if utils.ShouldSkipURL(url, as.config.SkipPatterns) {
			continue
		}

		as.session.MarkVisited(normalizedURL)

		wg.Add(1)
		go func(pageURL string, pageDepth int) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := as.browserService.CaptureScreenshot(pageURL)
			resultsChan <- result

			if result.Success && pageDepth < as.config.MaxDepth {
				links, err := as.browserService.ExtractLinks(pageURL)
				if err == nil {
					as.addNewLinksToQueue(links, pageDepth+1)
				}
			}
		}(url, depth)
	}

	wg.Wait()
	close(resultsChan)

	return nil
}

// addnewlinkstoqueue adds valid, unvisited links to the crawl queue at the given depth
func (as *AppService) addNewLinksToQueue(links []string, depth int) {
	for _, link := range links {
		fixedLink := utils.FixRelativeURL(link, as.config.BaseURL)
		if utils.IsValidURL(fixedLink, as.config.BaseURL) {
			normalizedLink := utils.NormalizeURL(fixedLink)
			if !as.session.IsVisited(normalizedLink) && !as.session.IsExisting(normalizedLink) {
				if !utils.ShouldSkipURL(fixedLink, as.config.SkipPatterns) {
					as.session.AddURL(fixedLink, depth)
				}
			}
		}
	}
}

// generatereport loads existing report, generates new report with session data, logs stats
func (as *AppService) GenerateReport() error {
	log.Printf("\033[36m> Generating reports...\033[0m")

	existingReport, err := as.reportService.LoadExistingReport()
	if err != nil {
		log.Printf("\033[36m> No existing report found, creating new one\033[0m")
		existingReport = nil
	}

	if err := as.reportService.GenerateReport(as.session, existingReport); err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}

	total, success, failed := as.session.GetStats()
	elapsed := as.session.GetElapsedTime()

	log.Printf("\033[32m> Report generated successfully\033[0m")
	log.Printf("\033[36m> Session Stats:\033[0m")
	log.Printf("\033[36m> Total: %d pages\033[0m", total)
	log.Printf("\033[32m> Success: %d pages\033[0m", success)
	log.Printf("\033[31m> Failed: %d pages\033[0m", failed)
	log.Printf("\033[36m> Duration: %.2f seconds\033[0m", elapsed.Seconds())

	return nil
}

// cleanup closes browser service and logs completion
func (as *AppService) Cleanup() {
	log.Printf("\033[36m> Cleaning up resources...\033[0m")

	if as.browserService != nil {
		as.browserService.Close()
	}

	log.Printf("\033[32m> Cleanup complete\033[0m")
}

// run orchestrates the entire application flow, initialize, discover, crawl, report, cleanup
func (as *AppService) Run() error {
	defer as.Cleanup()

	if err := as.Initialize(); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	if err := as.DiscoverURLs(); err != nil {
		return fmt.Errorf("URL discovery failed: %w", err)
	}

	if err := as.CrawlWebsite(); err != nil {
		return fmt.Errorf("website crawl failed: %w", err)
	}

	if err := as.GenerateReport(); err != nil {
		return fmt.Errorf("report generation failed: %w", err)
	}

	log.Printf("\033[32m> Framely completed successfully!\033[0m")
	return nil
}
