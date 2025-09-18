package models

import "time"

// screenshotresult represents the result of a screenshot capture operation
type ScreenshotResult struct {
	URL       string    `json:"url"`
	Filename  string    `json:"filename"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	FileSize  int64     `json:"fileSize,omitempty"`
	Duration  int64     `json:"duration,omitempty"`
}

// report represents the overall report of a crawl session
type Report struct {
	BaseURL               string             `json:"baseUrl"`
	TotalPages            int                `json:"totalPages"`
	SuccessfulScreenshots int                `json:"successfulScreenshots"`
	FailedScreenshots     int                `json:"failedScreenshots"`
	Timestamp             time.Time          `json:"timestamp"`
	LastUpdate            *time.Time         `json:"lastUpdate,omitempty"`
	NewPagesInThisRun     int                `json:"newPagesInThisRun"`
	TotalDuration         int64              `json:"totalDuration"`
	AveragePageSize       int64              `json:"averagePageSize"`
	Results               []ScreenshotResult `json:"results"`
}

// sitemapurl represents a url entry in a sitemap
type SitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

// urlset represents a collection of sitemap urls
type URLSet struct {
	URLs []SitemapURL `xml:"url"`
}

// crawlsession manages the state of a website crawl
type CrawlSession struct {
	baseURL        string
	visitedURLs    map[string]bool
	discoveredURLs map[string]bool
	existingURLs   map[string]bool
	urlQueue       []string
	depthMap       map[string]int
	results        []ScreenshotResult
	startTime      time.Time
}

// newcrawlsession creates a new crawlsession with initialized maps and start time
func NewCrawlSession(baseURL string) *CrawlSession {
	return &CrawlSession{
		baseURL:        baseURL,
		visitedURLs:    make(map[string]bool),
		discoveredURLs: make(map[string]bool),
		existingURLs:   make(map[string]bool),
		urlQueue:       make([]string, 0),
		depthMap:       make(map[string]int),
		results:        make([]ScreenshotResult, 0),
		startTime:      time.Now(),
	}
}

// addurl adds a url to the queue if not visited or existing, with the given depth
func (cs *CrawlSession) AddURL(url string, depth int) {
	if !cs.visitedURLs[url] && !cs.existingURLs[url] {
		cs.urlQueue = append(cs.urlQueue, url)
		cs.depthMap[url] = depth
	}
}

// markvisited marks a url as visited
func (cs *CrawlSession) MarkVisited(url string) {
	cs.visitedURLs[url] = true
}

// markexisting marks a url as existing
func (cs *CrawlSession) MarkExisting(url string) {
	cs.existingURLs[url] = true
}

// addresult adds a screenshot result to the session
func (cs *CrawlSession) AddResult(result ScreenshotResult) {
	cs.results = append(cs.results, result)
}

// getnexturl returns the next url from the queue, its depth, and if there is one
func (cs *CrawlSession) GetNextURL() (string, int, bool) {
	if len(cs.urlQueue) == 0 {
		return "", 0, false
	}

	url := cs.urlQueue[0]
	cs.urlQueue = cs.urlQueue[1:]
	depth := cs.depthMap[url]

	return url, depth, true
}

// isvisited checks if a url has been visited
func (cs *CrawlSession) IsVisited(url string) bool {
	return cs.visitedURLs[url]
}

// isexisting checks if a url is marked as existing
func (cs *CrawlSession) IsExisting(url string) bool {
	return cs.existingURLs[url]
}

// getresults returns all screenshot results
func (cs *CrawlSession) GetResults() []ScreenshotResult {
	return cs.results
}

// getelapsedtime returns the time elapsed since session start
func (cs *CrawlSession) GetElapsedTime() time.Duration {
	return time.Since(cs.startTime)
}

// getstats returns total, success, and failed counts from results
func (cs *CrawlSession) GetStats() (total int, success int, failed int) {
	total = len(cs.results)
	for _, result := range cs.results {
		if result.Success {
			success++
		} else {
			failed++
		}
	}
	return
}
