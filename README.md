![Banner](https://www.upload.ee/image/18611472/framely-banner.png)

# Framely

**Capture the Web, Frame by Frame**

Framely is a tool developed to automatically take screenshots of websites. It is written in Go and uses the Chrome headless browser to produce high-quality screenshots.

## Features

- **Automatic Crawling**: Deeply crawls the website and discovers all pages
- **Parallel Processing**: Takes screenshots in parallel with multiple workers
- **Smart Filtering**: Skips unnecessary files (PDFs, images, etc.)
- **Sitemap and Robots.txt Support**: Discovers additional URLs
- **Detailed Reporting**: Generates reports in JSON and text formats
- **Configuration**: Customizable with flexible settings
- **Security**: Domain verification and input validation

## Installation

### Requirements

- Go 1.24.3 or higher
- Google Chrome or Chromium browser

### Steps

> **Note:** A pre-built `framely` binary is available in the repository. You can run this file directly, or follow the steps below to build it yourself.

1. Clone the project:
   ```bash
   git clone https://github.com/Queaxtra/framely.git
   cd framely
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build:
   ```bash
   go build -o framely src/main.go
   ```

## Usage

Basic usage:
```bash
./framely
```

When you run the tool, interactive mode will start and you can configure the following settings:

- Target website URL
- Sitemap check
- Robots.txt check
- Maximum crawl depth
- Number of parallel workers
- URL patterns to skip

### Examples

- Simple usage: Just enter the URL and use default settings
- Deep scan: Increase the maximum depth
- Fast scan: Increase the number of parallel workers

## Configuration

The tool supports the following settings:

- **Maximum Depth**: Crawl depth (default: 5)
- **Parallel Workers**: Number of pages processed simultaneously (default: 5)
- **Screenshot Delay**: Wait time before capturing the page
- **Viewport Size**: Screenshot dimensions
- **Quality**: JPEG quality (1-100)
- **Skip Patterns**: Patterns to skip specific URLs

## Outputs

The tool generates the following files:

- `screenshots/`: All screenshots
- `report.json`: Detailed JSON report
- `summary.txt`: Summary text report

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to your branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.

## Contact

If you have any issues regarding the project, please send an email to `fatih@etik.com` or open an issue on GitHub.
