package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"

	"framely/src/config"
	"framely/src/models"
	"framely/src/utils"
)

// browserservice holds config, context, and cancel for browser operations
type BrowserService struct {
	config *config.Config
	ctx    context.Context
	cancel context.CancelFunc
}

// newbrowserservice creates a new browserservice instance with chromedp setup
func NewBrowserService(cfg *config.Config) *BrowserService {
	opts := make([]chromedp.ExecAllocatorOption, 0, len(chromedp.DefaultExecAllocatorOptions)+len(config.CHROME_FLAGS)+1)
	opts = append(opts, chromedp.DefaultExecAllocatorOptions[:]...)

	for _, flag := range config.CHROME_FLAGS {
		opts = append(opts, chromedp.Flag(strings.TrimPrefix(flag, "--"), true))
	}

	opts = append(opts, chromedp.UserAgent(cfg.UserAgent))

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx)

	return &BrowserService{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}
}

// close cancels the browser context
func (bs *BrowserService) Close() {
	if bs.cancel != nil {
		bs.cancel()
	}
}

// capturescreenshot navigates to the url, takes a full screenshot, saves it to file, returns result
func (bs *BrowserService) CaptureScreenshot(url string) models.ScreenshotResult {
	startTime := time.Now()
	filename := utils.GenerateFilename(url)
	filepath := filepath.Join(bs.config.OutputDir, filename)

	log.Printf("\033[36m> Capturing screenshot: %s\033[0m", url)

	result := models.ScreenshotResult{
		URL:       url,
		Filename:  filename,
		Timestamp: startTime,
	}

	var screenshotData []byte
	err := chromedp.Run(bs.ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(time.Duration(bs.config.ScreenshotDelay)*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return emulation.SetDeviceMetricsOverride(
				int64(bs.config.ViewportWidth),
				int64(bs.config.ViewportHeight),
				1, false,
			).Do(ctx)
		}),
		chromedp.FullScreenshot(&screenshotData, bs.config.Quality),
	)

	duration := time.Since(startTime).Milliseconds()
	result.Duration = duration

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		log.Printf("\033[31m> Screenshot failed for %s: %s\033[0m", url, err.Error())
		return result
	}

	if err := os.WriteFile(filepath, screenshotData, 0644); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("File write error: %s", err.Error())
		log.Printf("\033[31m> File write failed for %s: %s\033[0m", filename, err.Error())
		return result
	}

	if fileInfo, err := os.Stat(filepath); err == nil {
		result.FileSize = fileInfo.Size()
	}

	result.Success = true
	log.Printf("\033[32m> Screenshot saved: %s (%.2fKB)\033[0m", filename, float64(result.FileSize)/1024)

	return result
}

// extractlinks navigates to the url, extracts all links, filters and normalizes valid ones
func (bs *BrowserService) ExtractLinks(url string) ([]string, error) {
	var links []string

	err := chromedp.Run(bs.ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(time.Duration(bs.config.ScreenshotDelay)*time.Second),
		chromedp.Evaluate(`
			Array.from(document.querySelectorAll('a[href]')).map(link => {
				try {
					return new URL(link.href, window.location.href).href;
				} catch (e) {
					return null;
				}
			}).filter(href => href !== null);
		`, &links),
	)

	if err != nil {
		log.Printf("\033[31m> Link extraction failed for %s: %s\033[0m", url, err.Error())
		return nil, err
	}

	validLinks := make([]string, 0)
	seenLinks := make(map[string]bool)

	for _, link := range links {
		if utils.IsValidURL(link, bs.config.BaseURL) && !seenLinks[link] {
			normalizedLink := utils.NormalizeURL(link)
			if !seenLinks[normalizedLink] {
				validLinks = append(validLinks, link)
				seenLinks[normalizedLink] = true
			}
		}
	}

	log.Printf("\033[32m> Extracted %d valid links from %s\033[0m", len(validLinks), url)
	return validLinks, nil
}

// testconnection navigates to the url and checks if the page loads successfully
func (bs *BrowserService) TestConnection(url string) error {
	err := chromedp.Run(bs.ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
	)

	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	log.Printf("\033[32m> Connection test successful for %s\033[0m", url)
	return nil
}
