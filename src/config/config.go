package config

const (
	SCREENSHOTS_DIR   = "screenshots"
	REPORT_FILE      = "report.json"
	SUMMARY_FILE     = "summary.txt"
	DEFAULT_MAX_DEPTH = 5
	DEFAULT_PARALLEL_WORKERS = 5
	DEFAULT_SCREENSHOT_DELAY = 3
	DEFAULT_REQUEST_DELAY = 1
	DEFAULT_VIEWPORT_WIDTH = 1920
	DEFAULT_VIEWPORT_HEIGHT = 1080
	DEFAULT_SCREENSHOT_QUALITY = 90
	DOMAIN_REGEX = `^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	)

var (
	EXCLUDED_EXTENSIONS = []string{
		".pdf", ".doc", ".docx", ".xls", ".xlsx",
		".zip", ".rar", ".exe", ".dmg", ".pkg",
		".mp4", ".avi", ".mov", ".mp3", ".wav",
		".jpg", ".jpeg", ".png", ".gif", ".svg",
	}

	DEFAULT_SKIP_PATTERNS = []string{
		"/wp-admin",
	}

	CHROME_FLAGS = []string{
		"--headless",
		"--no-sandbox",
		"--disable-setuid-sandbox",
		"--disable-dev-shm-usage",
		"--disable-accelerated-2d-canvas",
		"--no-first-run",
		"--no-zygote",
		"--disable-gpu",
		"--disable-extensions",
		"--disable-plugins",
		"--disable-images",
		"--disable-javascript",
	}

	DEFAULT_USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"
)

// config holds all configuration settings for the application
type Config struct {
	BaseURL          string
	MaxDepth         int
	ParallelWorkers  int
	ScreenshotDelay  int
	RequestDelay     int
	ViewportWidth    int
	ViewportHeight   int
	Quality          int
	CheckSitemap     bool
	CheckRobots      bool
	SkipPatterns     []string
	UserAgent        string
	OutputDir        string
}

// newconfig creates a new config instance with default values and the given baseurl
func NewConfig(baseURL string) *Config {
	return &Config{
		BaseURL:         baseURL,
		MaxDepth:        DEFAULT_MAX_DEPTH,
		ParallelWorkers: DEFAULT_PARALLEL_WORKERS,
		ScreenshotDelay: DEFAULT_SCREENSHOT_DELAY,
		RequestDelay:    DEFAULT_REQUEST_DELAY,
		ViewportWidth:   DEFAULT_VIEWPORT_WIDTH,
		ViewportHeight:  DEFAULT_VIEWPORT_HEIGHT,
		Quality:         DEFAULT_SCREENSHOT_QUALITY,
		CheckSitemap:    true,
		CheckRobots:     true,
		SkipPatterns:    DEFAULT_SKIP_PATTERNS,
		UserAgent:       DEFAULT_USER_AGENT,
		OutputDir:       SCREENSHOTS_DIR,
	}
}
