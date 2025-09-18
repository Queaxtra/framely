package services

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"framely/src/models"
	"framely/src/utils"
)

// discoveryservice holds baseurl and httpclient for url discovery operations
type DiscoveryService struct {
	baseURL    string
	httpClient *http.Client
}

// newdiscoveryservice creates a new discoveryservice instance with the given baseurl
func NewDiscoveryService(baseURL string) *DiscoveryService {
	return &DiscoveryService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// discoverurls discovers urls from sitemap and robots.txt based on flags, normalizes and validates them
func (ds *DiscoveryService) DiscoverURLs(checkSitemap, checkRobots bool) []string {
	discoveredURLs := make(map[string]bool)

	if checkSitemap {
		sitemapURLs := ds.parseSitemap()
		for _, url := range sitemapURLs {
			normalizedURL := utils.NormalizeURL(url)
			discoveredURLs[normalizedURL] = true
		}
		log.Printf("\033[32m> Sitemap discovery: %d URLs found\033[0m", len(sitemapURLs))
	}

	if checkRobots {
		robotsURLs := ds.parseRobotsTxt()
		for _, url := range robotsURLs {
			normalizedURL := utils.NormalizeURL(url)
			discoveredURLs[normalizedURL] = true
		}
		log.Printf("\033[32m> Robots.txt discovery: %d URLs found\033[0m", len(robotsURLs))
	}

	urls := make([]string, 0, len(discoveredURLs))
	for url := range discoveredURLs {
		fullURL := utils.FixRelativeURL(url, ds.baseURL)
		if utils.IsValidURL(fullURL, ds.baseURL) {
			urls = append(urls, fullURL)
		}
	}

	return urls
}

// parsesitemap fetches and parses the main sitemap.xml for urls
func (ds *DiscoveryService) parseSitemap() []string {
	return ds.fetchAndParseSitemap(ds.baseURL + "/sitemap.xml")
}

// parserobotstxt fetches robots.txt, extracts sitemap references, and parses those sitemaps for urls
func (ds *DiscoveryService) parseRobotsTxt() []string {
	robotsURL := ds.baseURL + "/robots.txt"

	resp, err := ds.httpClient.Get(robotsURL)
	if err != nil {
		log.Printf("\033[31m> Robots.txt fetch error: %s\033[0m", err.Error())
		return []string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("\033[31m> Robots.txt not accessible: HTTP %d\033[0m", resp.StatusCode)
		return []string{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("\033[31m> Robots.txt read error: %s\033[0m", err.Error())
		return []string{}
	}

	sitemapURLs := ds.extractSitemapReferences(string(body))
	allURLs := make([]string, 0)

	for _, sitemapURL := range sitemapURLs {
		urls := ds.fetchAndParseSitemap(sitemapURL)
		allURLs = append(allURLs, urls...)
	}

	return allURLs
}

// extractsitemapreferences extracts sitemap urls from robots.txt content
func (ds *DiscoveryService) extractSitemapReferences(robotsContent string) []string {
	lines := strings.Split(robotsContent, "\n")
	sitemapURLs := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "sitemap:") {
			sitemapURL := strings.TrimSpace(strings.TrimPrefix(line, "Sitemap:"))
			sitemapURL = strings.TrimSpace(strings.TrimPrefix(sitemapURL, "sitemap:"))

			if strings.HasPrefix(sitemapURL, "http") {
				sitemapURLs = append(sitemapURLs, sitemapURL)
			}
		}
	}

	return sitemapURLs
}

// fetchandparsesitemap fetches a sitemap url and parses it for valid urls
func (ds *DiscoveryService) fetchAndParseSitemap(sitemapURL string) []string {
	resp, err := ds.httpClient.Get(sitemapURL)
	if err != nil {
		log.Printf("\033[31m> Sitemap fetch error (%s): %s\033[0m", sitemapURL, err.Error())
		return []string{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("\033[31m> Sitemap not accessible (%s): HTTP %d\033[0m", sitemapURL, resp.StatusCode)
		return []string{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("\033[31m> Sitemap read error (%s): %s\033[0m", sitemapURL, err.Error())
		return []string{}
	}

	var urlset models.URLSet
	if err := xml.Unmarshal(body, &urlset); err != nil {
		log.Printf("\033[31m> Sitemap parse error (%s): %s\033[0m", sitemapURL, err.Error())
		return []string{}
	}

	urls := make([]string, 0, len(urlset.URLs))
	for _, url := range urlset.URLs {
		if url.Loc != "" && utils.IsValidURL(url.Loc, ds.baseURL) {
			urls = append(urls, url.Loc)
		}
	}

	return urls
}

// testsitemapaccess tests if sitemap.xml is accessible
func (ds *DiscoveryService) TestSitemapAccess() error {
	sitemapURL := ds.baseURL + "/sitemap.xml"

	resp, err := ds.httpClient.Get(sitemapURL)
	if err != nil {
		return fmt.Errorf("sitemap test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sitemap not accessible: HTTP %d", resp.StatusCode)
	}

	return nil
}

// testrobotsaccess tests if robots.txt is accessible
func (ds *DiscoveryService) TestRobotsAccess() error {
	robotsURL := ds.baseURL + "/robots.txt"

	resp, err := ds.httpClient.Get(robotsURL)
	if err != nil {
		return fmt.Errorf("robots.txt test failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("robots.txt not accessible: HTTP %d", resp.StatusCode)
	}

	return nil
}
