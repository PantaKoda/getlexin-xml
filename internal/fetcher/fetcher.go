package fetcher

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"getlexin-xml/internal/models"
	"getlexin-xml/internal/parser"
)

// DownloadManager handles concurrent downloads
type DownloadManager struct {
	Concurrency int
	OutputDir   string
	Results     chan models.DownloadResult
}

// NewDownloadManager creates a new download manager
func NewDownloadManager(concurrency int, outputDir string) *DownloadManager {
	return &DownloadManager{
		Concurrency: concurrency,
		OutputDir:   outputDir,
		Results:     make(chan models.DownloadResult),
	}
}

// StartDownloads begins downloading the selected directories concurrently
func (dm *DownloadManager) StartDownloads(directories []models.Directory) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, dm.Concurrency)

	// Create output directory if it doesn't exist
	err := os.MkdirAll(dm.OutputDir, 0755)
	if err != nil {
		log.Printf("Failed to create output directory: %v", err)
		close(dm.Results)
		return
	}

	// Start a goroutine to close the results channel once all downloads are done
	go func() {
		wg.Wait()
		close(dm.Results)
	}()

	// Start download goroutines
	for _, dir := range directories {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(dir models.Directory) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			result := dm.downloadDirectory(dir)
			dm.Results <- result
		}(dir)
	}
}

// downloadDirectory downloads all XML files from a language directory
func (dm *DownloadManager) downloadDirectory(dir models.Directory) models.DownloadResult {
	result := models.DownloadResult{
		Directory: dir,
		Success:   false,
	}

	// Create directory-specific output folder
	dirPath := filepath.Join(dm.OutputDir, dir.Code)
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		result.Error = fmt.Errorf("failed to create directory %s: %v", dirPath, err)
		return result
	}

	// Fetch the directory page
	resp, err := http.Get(dir.URL)
	if err != nil {
		result.Error = fmt.Errorf("failed to fetch directory: %v", err)
		return result
	}
	defer resp.Body.Close()

	// Read the directory index content
	indexContent, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Errorf("failed to read directory page: %v", err)
		return result
	}

	// Save the directory index
	err = os.WriteFile(filepath.Join(dirPath, "index.html"), indexContent, 0644)
	if err != nil {
		result.Error = fmt.Errorf("failed to save index file: %v", err)
		return result
	}

	// Parse the directory contents to find XML files
	xmlFiles, err := parser.ParseDirectoryContents(string(indexContent))
	if err != nil {
		result.Error = fmt.Errorf("failed to parse directory contents: %v", err)
		return result
	}

	// Download each XML file
	var totalBytes int64
	for _, file := range xmlFiles {
		fileURL := dir.URL + file.Href
		filePath := filepath.Join(dirPath, file.Name)

		bytes, err := dm.downloadFile(filePath, fileURL)
		if err != nil {
			log.Printf("Error downloading %s: %v", file.Name, err)
		} else {
			result.FileCount++
			totalBytes += bytes
		}
	}

	// Create metadata file
	metaFile, err := os.Create(filepath.Join(dirPath, "metadata.txt"))
	if err != nil {
		result.Error = fmt.Errorf("failed to create metadata file: %v", err)
		return result
	}
	defer metaFile.Close()

	// Write metadata
	_, err = fmt.Fprintf(metaFile, "Code: %s\nName: %s\nURL: %s\nDownloaded: %s\nFiles: %d\nTotal Size: %d bytes\n",
		dir.Code, dir.Name, dir.URL, time.Now().Format(time.RFC3339), result.FileCount, totalBytes)
	if err != nil {
		result.Error = fmt.Errorf("failed to write metadata: %v", err)
		return result
	}

	result.Success = true
	result.TotalBytes = totalBytes
	return result
}

// downloadFile downloads a file from a URL to a local path and returns the number of bytes downloaded
func (dm *DownloadManager) downloadFile(filepath string, url string) (int64, error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad status: %s", resp.Status)
	}

	// Write the body to file
	bytesWritten, err := io.Copy(out, resp.Body)
	if err != nil {
		return 0, err
	}

	return bytesWritten, nil
}
