package utils

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"framely/src/config"
)

// isvalidurl checks if the given url is valid, it parses the url, compares host and scheme with baseurl, checks for excluded extensions, excludes mailto, tel, javascript protocols
func IsValidURL(urlStr, baseURL string) bool {
	if urlStr == "" {
		return false
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	baseU, err := url.Parse(baseURL)
	if err != nil {
		return false
	}

	if u.Host != baseU.Host || u.Scheme != baseU.Scheme {
		return false
	}

	pathname := strings.ToLower(u.Path)
	for _, ext := range config.EXCLUDED_EXTENSIONS {
		if strings.HasSuffix(pathname, ext) {
			return false
		}
	}

	if strings.Contains(urlStr, "mailto:") || strings.Contains(urlStr, "tel:") || strings.Contains(urlStr, "javascript:") {
		return false
	}

	return true
}

// normalizeurl normalizes the url by removing fragment and query, setting path to / if empty, trimming trailing slash
func NormalizeURL(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	u.Fragment = ""
	u.RawQuery = ""

	if u.Path == "" {
		u.Path = "/"
	}

	u.Path = strings.TrimSuffix(u.Path, "/")
	if u.Path == "" {
		u.Path = "/"
	}

	return u.String()
}

// fixrelativeurl fixes relative urls by resolving them against the baseurl, if already absolute, returns as is
func FixRelativeURL(link, baseURL string) string {
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return link
	}

	baseU, err := url.Parse(baseURL)
	if err != nil {
		return link
	}

	linkU, err := url.Parse(link)
	if err != nil {
		return link
	}

	resolvedURL := baseU.ResolveReference(linkU)
	return resolvedURL.String()
}

// shouldskipurl checks if the url should be skipped based on the skip patterns, compares case-insensitively
func ShouldSkipURL(urlStr string, skipPatterns []string) bool {
	for _, pattern := range skipPatterns {
		if strings.Contains(strings.ToLower(urlStr), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// generatefilename generates a filename from the url, replaces path slashes with underscores, adds query if present, sanitizes, appends .png
func GenerateFilename(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return generateTimestampFilename()
	}

	filename := strings.ReplaceAll(u.Path, "/", "_")

	if filename == "" || filename == "_" {
		filename = "homepage"
	}

	if u.RawQuery != "" {
		queryPart := strings.ReplaceAll(u.RawQuery, "=", "_")
		queryPart = strings.ReplaceAll(queryPart, "&", "_")
		filename += "_" + queryPart
	}

	filename = sanitizeFilename(filename)

	if filename == "" {
		filename = generateTimestampFilename()
	}

	return filename + ".png"
}

// sanitizefilename sanitizes the filename by replacing invalid chars with underscore, collapsing multiple underscores, trimming, limiting length
func sanitizeFilename(filename string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	filename = reg.ReplaceAllString(filename, "_")

	reg = regexp.MustCompile(`_+`)
	filename = reg.ReplaceAllString(filename, "_")

	filename = strings.Trim(filename, "_")

	if len(filename) > 200 {
		filename = filename[:200]
	}

	return filename
}

// generatetimestampfilename generates a timestamp-based filename for fallback
func generateTimestampFilename() string {
	return "page_" + time.Now().Format("20060102_150405")
}

// extractdomain extracts the domain from the url
func ExtractDomain(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Host
}

// ishttps checks if the url uses https scheme
func IsHTTPS(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme == "https"
}

// addhttpsifmissing adds https prefix if the url does not have http or https
func AddHTTPSIfMissing(urlStr string) string {
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return "https://" + urlStr
	}
	return urlStr
}

// getpathfromurl extracts the path from the url
func GetPathFromURL(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Path
}

// hasqueryparams checks if the url has query parameters
func HasQueryParams(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return u.RawQuery != ""
}
