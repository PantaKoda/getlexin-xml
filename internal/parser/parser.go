package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"

	"getlexin-xml/internal/models"
)

// FetchDirectories fetches and parses the directory list from the lexin site
func FetchDirectories(baseURL string) ([]models.Directory, error) {
	// Fetch the XML from the URL
	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	xmlData := string(body)

	// Clean up the XML to remove the DOCTYPE declaration which can cause parsing issues
	xmlData = RemoveDOCTYPE(xmlData)

	// Parse the XML
	var svn models.SVN
	err = xml.Unmarshal([]byte(xmlData), &svn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse XML: %v", err)
	}

	// Extract directories with translated names
	var directories []models.Directory
	for _, dir := range svn.Index.Dirs {
		langName, ok := models.LanguageMap[dir.Name]
		if !ok {
			langName = strings.Title(dir.Name) // Capitalize if not in map
		}

		directories = append(directories, models.Directory{
			Code:        dir.Name,
			Name:        langName,
			URL:         baseURL + dir.Href,
			Description: fmt.Sprintf("Swedish-%s lexicon", langName),
			Selected:    false,
		})
	}

	return directories, nil
}

// ParseDirectoryContents parses a directory's XML content and returns XML filenames
func ParseDirectoryContents(xmlContent string) ([]models.File, error) {
	// Clean up the XML
	xmlContent = RemoveDOCTYPE(xmlContent)

	// Parse the XML
	var svn models.SVN
	err := xml.Unmarshal([]byte(xmlContent), &svn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory XML: %v", err)
	}

	// Filter files to include only XML files
	var xmlFiles []models.File
	for _, file := range svn.Index.Files {
		if strings.HasSuffix(file.Href, ".xml") {
			xmlFiles = append(xmlFiles, file)
		}
	}

	return xmlFiles, nil
}

// RemoveDOCTYPE removes DOCTYPE declaration from XML string
func RemoveDOCTYPE(xmlStr string) string {
	docTypeStart := strings.Index(xmlStr, "<!DOCTYPE")
	if docTypeStart == -1 {
		return xmlStr
	}

	docTypeEnd := strings.Index(xmlStr[docTypeStart:], ">")
	if docTypeEnd == -1 {
		return xmlStr
	}

	return xmlStr[:docTypeStart] + xmlStr[docTypeStart+docTypeEnd+1:]
}
