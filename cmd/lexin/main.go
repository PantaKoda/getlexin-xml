package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"

	"getlexin-xml/internal/fetcher"
	"getlexin-xml/internal/models"
	"getlexin-xml/internal/parser"
	"getlexin-xml/internal/ui"
)

func main() {
	// Define command line flags
	outputDir := flag.String("out", "lexin_downloads", "Output directory for downloads")
	concurrency := flag.Int("concurrency", 3, "Number of concurrent downloads")
	flag.Parse()

	// URL to fetch
	baseURL := "https://sprakresurser.isof.se/lexin/"

	// Fetch the directories from the URL
	directories, err := parser.FetchDirectories(baseURL)
	if err != nil {
		log.Fatalf("Failed to fetch directories: %v", err)
	}

	// Interactive TUI interface
	p := tea.NewProgram(ui.NewModel(directories, baseURL, *outputDir, *concurrency))

	result, err := p.Run()
	if err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}

	// Process the TUI result
	m, ok := result.(ui.Model)
	if !ok {
		fmt.Println("Error: couldn't process UI model")
		return
	}

	// Exit if not in download mode
	if !m.ShowDownloads {
		fmt.Println("Exiting without downloading.")
		return
	}

	// Get selected directories
	selectedDirs := m.GetSelectedDirectories(directories)
	if len(selectedDirs) == 0 {
		fmt.Println("No languages selected. Exiting.")
		return
	}

	// Start the download process
	fmt.Printf("\nStarting download of %d language directories...\n\n", len(selectedDirs))
	err = downloadWithProgressReporting(selectedDirs, *outputDir, *concurrency)
	if err != nil {
		log.Fatalf("Error during download: %v", err)
	}

	fmt.Println("\nAll downloads complete!")
}

// downloadWithProgressReporting handles the downloads and displays progress
func downloadWithProgressReporting(directories []models.Directory, outputDir string, concurrency int) error {
	// Create output directory if it doesn't exist
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}
	// Create download manager
	downloadManager := fetcher.NewDownloadManager(concurrency, outputDir)

	// Start downloads in background
	downloadManager.StartDownloads(directories)

	// Track progress
	startTime := time.Now()
	completed := 0
	total := len(directories)

	// Print header
	fmt.Printf("%-15s %-20s %-10s %-10s\n", "LANGUAGE", "STATUS", "FILES", "SIZE")
	fmt.Println(strings.Repeat("-", 60))

	// Process results as they come in
	var lastError error
	for result := range downloadManager.Results {
		completed++
		status := "ERROR"
		if result.Success {
			status = "COMPLETED"
		} else if result.Error != nil {
			lastError = result.Error
		}

		sizeStr := "N/A"
		if result.Success {
			sizeStr = formatBytes(result.TotalBytes)
		}

		// Print the result
		fmt.Printf("%-15s %-20s %-10d %-10s [%d/%d]\n",
			result.Directory.Code,
			status,
			result.FileCount,
			sizeStr,
			completed,
			total)
	}

	// Print summary
	elapsed := time.Since(startTime)
	fmt.Printf("\nDownload summary:\n")
	fmt.Printf("- Languages processed: %d\n", completed)
	fmt.Printf("- Time elapsed: %s\n", elapsed.Round(time.Second))

	return lastError
}

// formatBytes converts bytes to human readable string using go-humanize
func formatBytes(bytes int64) string {
	return humanize.Bytes(uint64(bytes))
}
