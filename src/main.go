package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"framely/src/config"
	"framely/src/services"
	"framely/src/utils"
)

// main is the entry point of the application, it clears the screen, prints the banner,
// collects user input for configuration, initializes the app service, and runs it
func main() {
	clearScreen()
	printBanner()

	cfg, err := collectUserInput()
	if err != nil {
		log.Fatalf("\033[31m> Configuration error: %v\033[0m", err)
	}

	appService := services.NewAppService(cfg)

	if err := appService.Run(); err != nil {
		log.Fatalf("\033[31m> Application error: %v\033[0m", err)
	}
}

// clearscreen clears the terminal screen, it attempts to use 'clear' for unix-like systems,
// and falls back to 'cls' for windows if the first command fails
func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	if cmd.Run() != nil {
		cmd = exec.Command("cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// printbanner displays the application's ascii art banner, version information,
// and github link to the user
func printBanner() {
	fmt.Printf("\033]0;Framely\007")
	fmt.Println("▗▄▄▄▖▗▄▄▖  ▗▄▖ ▗▖  ▗▖▗▄▄▄▖▗▖ ▗▖  ▗▖")
	fmt.Println("▐▌   ▐▌ ▐▌▐▌ ▐▌▐▛▚▞▜▌▐▌   ▐▌  ▝▚▞▘ ")
	fmt.Println("▐▛▀▀▘▐▛▀▚▖▐▛▀▜▌▐▌  ▐▌▐▛▀▀▘▐▌   ▐▌  ")
	fmt.Println("▐▌   ▐▌ ▐▌▐▌ ▐▌▐▌  ▐▌▐▙▄▄▖▐▙▄▄▖▐▌  ")
	fmt.Println("                                       ")
	fmt.Printf("\033[32mCapture the Web, Frame by Frame\033[0m\n")
	fmt.Println()
	fmt.Printf("\033[34mv0.1 - ALPHA\033[0m\n")
	fmt.Printf("\033[33mGithub: https://github.com/Queaxtra/framely\033[0m\n")
	fmt.Printf("\033[33m===========================================\033[0m\n")
	fmt.Println()
}

// collectuserinput prompts the user for all necessary configuration options,
// including target url, sitemap checking, robots.txt checking, crawl depth,
// parallel processing, and skip patterns, it returns a fully configured config struct
func collectUserInput() (*config.Config, error) {
	reader := bufio.NewReader(os.Stdin)

	baseURL, err := getTargetURL(reader)
	if err != nil {
		return nil, err
	}

	cfg := config.NewConfig(baseURL)

	if err := configureSitemap(reader, cfg); err != nil {
		return nil, err
	}

	if err := configureRobots(reader, cfg); err != nil {
		return nil, err
	}

	if err := configureDepth(reader, cfg); err != nil {
		return nil, err
	}

	if err := configureParallel(reader, cfg); err != nil {
		return nil, err
	}

	if err := configureSkipPatterns(reader, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// gettargeturl reads and validates the target website url from user input,
// it ensures the url is not empty, adds https if missing, parses it, and checks
// the domain against a regex pattern for validity
func getTargetURL(reader *bufio.Reader) (string, error) {
	input, err := readInput(reader, "\033[36m> Enter target website URL: \033[0m")
	if err != nil {
		return "", fmt.Errorf("failed to read URL: %w", err)
	}

	targetURL := strings.TrimSpace(input)
	if targetURL == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	targetURL = utils.AddHTTPSIfMissing(targetURL)

	u, err := url.Parse(targetURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL format: %w", err)
	}

	host := u.Host
	if host == "" {
		return "", fmt.Errorf("URL must have a host")
	}

	matched, err := regexp.MatchString(config.DOMAIN_REGEX, host)
	if err != nil {
		return "", fmt.Errorf("regex error: %w", err)
	}
	if !matched {
		return "", fmt.Errorf("invalid domain format")
	}

	fmt.Printf("\033[32m> Target set: %s\n\n\033[0m", targetURL)
	return targetURL, nil
}

// configuresitemap prompts the user to enable or disable checking sitemap.xml
// for additional urls and updates the configuration accordingly
func configureSitemap(reader *bufio.Reader, cfg *config.Config) error {
	input, err := readInput(reader, "\033[36m> Check sitemap.xml for additional URLs? (Y/n): \033[0m")
	if err != nil {
		return fmt.Errorf("failed to read sitemap option: %w", err)
	}

	cfg.CheckSitemap = parseYesNo(input, true)
	if cfg.CheckSitemap {
		fmt.Println("\033[32m> Sitemap checking enabled\033[0m")
	}
	if !cfg.CheckSitemap {
		fmt.Println("\033[32m> Sitemap checking disabled\033[0m")
	}

	return nil
}

// configurerobots prompts the user to enable or disable checking robots.txt
// for sitemap references and updates the configuration accordingly
func configureRobots(reader *bufio.Reader, cfg *config.Config) error {
	input, err := readInput(reader, "\033[36m> Check robots.txt for sitemap references? (Y/n): \033[0m")
	if err != nil {
		return fmt.Errorf("failed to read robots option: %w", err)
	}

	cfg.CheckRobots = parseYesNo(input, true)
	if cfg.CheckRobots {
		fmt.Println("\033[32m> Robots.txt checking enabled\033[0m")
	}
	if !cfg.CheckRobots {
		fmt.Println("\033[32m> Robots.txt checking disabled\033[0m")
	}

	return nil
}

// configuredepth prompts the user for the maximum crawl depth, validates the input
// (must be between 1 and 10), and sets it in the configuration, defaults to a predefined value
func configureDepth(reader *bufio.Reader, cfg *config.Config) error {
	input, err := readInput(reader, fmt.Sprintf("\033[36m> Maximum crawl depth (default %d): \033[0m", config.DEFAULT_MAX_DEPTH))
	if err != nil {
		return fmt.Errorf("failed to read depth option: %w", err)
	}

	if input != "" {
		depth, err := strconv.Atoi(input)
		if err != nil {
			return fmt.Errorf("invalid depth value: %w", err)
		}
		if depth < 1 {
			return fmt.Errorf("depth must be at least 1")
		}
		if depth > 10 {
			return fmt.Errorf("depth cannot exceed 10 for safety")
		}
		cfg.MaxDepth = depth
	}

	fmt.Printf("\033[32m> Max depth set to: %d\n\033[0m", cfg.MaxDepth)
	return nil
}

// configureparallel prompts the user to enable parallel processing, if enabled,
// it calls configureworkercount to set the number of workers, otherwise, sets to sequential
func configureParallel(reader *bufio.Reader, cfg *config.Config) error {
	input, err := readInput(reader, "\033[36m> Enable parallel processing? (y/N): \033[0m")
	if err != nil {
		return fmt.Errorf("failed to read parallel option: %w", err)
	}

	enableParallel := parseYesNo(input, false)
	if enableParallel {
		if err := configureWorkerCount(reader, cfg); err != nil {
			return err
		}
	}
	if !enableParallel {
		cfg.ParallelWorkers = 1
		fmt.Println("\033[32m> Sequential processing enabled\033[0m")
	}

	return nil
}

// configureworkercount prompts the user for the number of parallel workers,
// validates the input (must be between 1 and 10), and sets it in the configuration,
// defaults to a predefined value
func configureWorkerCount(reader *bufio.Reader, cfg *config.Config) error {
	input, err := readInput(reader, fmt.Sprintf("\033[36m> Number of parallel workers (default %d): \033[0m", config.DEFAULT_PARALLEL_WORKERS))
	if err != nil {
		return fmt.Errorf("failed to read worker count: %w", err)
	}

	if input != "" {
		workers, err := strconv.Atoi(input)
		if err != nil {
			return fmt.Errorf("invalid worker count: %w", err)
		}
		if workers < 1 {
			return fmt.Errorf("worker count must be at least 1")
		}
		if workers > 10 {
			return fmt.Errorf("worker count cannot exceed 10 for safety")
		}
		cfg.ParallelWorkers = workers
	}

	fmt.Printf("\033[32m> Parallel processing enabled with %d workers\n\033[0m", cfg.ParallelWorkers)
	return nil
}

// configureskippatterns prompts the user for additional url patterns to skip,
// parses comma-separated input, trims spaces, and appends valid patterns to the config,
// if no input, uses default patterns
func configureSkipPatterns(reader *bufio.Reader, cfg *config.Config) error {
	input, err := readInput(reader, "\033[36m> Additional URL patterns to skip (comma-separated, optional): \033[0m")
	if err != nil {
		return fmt.Errorf("failed to read skip patterns: %w", err)
	}

	if input != "" {
		patterns := strings.Split(input, ",")
		for _, pattern := range patterns {
			trimmed := strings.TrimSpace(pattern)
			if trimmed != "" {
				cfg.SkipPatterns = append(cfg.SkipPatterns, trimmed)
			}
		}
		fmt.Printf("\033[32m> Added %d custom skip patterns\n\033[0m", len(patterns))
	}
	if input == "" {
		fmt.Printf("\033[32m> Using %d default skip patterns\n\033[0m", len(cfg.SkipPatterns))
	}

	fmt.Println()
	return nil
}

// readinput displays the given prompt and reads a line of input from the user,
// trimming any leading or trailing whitespace before returning
func readInput(reader *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

// parseyesno interprets a string input as a yes/no response, if defaultyes is true,
// it returns true unless input is "n" or "no", otherwise, returns true only if input is "y" or "yes"
func parseYesNo(input string, defaultYes bool) bool {
	response := strings.ToLower(input)
	if defaultYes {
		return !(response == "n" || response == "no")
	}
	return response == "y" || response == "yes"
}
