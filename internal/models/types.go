package models

import "encoding/xml"

// XML structure types
type SVN struct {
	XMLName xml.Name `xml:"svn"`
	Version string   `xml:"version,attr"`
	Href    string   `xml:"href,attr"`
	Index   Index    `xml:"index"`
}

type Index struct {
	XMLName xml.Name `xml:"index"`
	Rev     string   `xml:"rev,attr"`
	Path    string   `xml:"path,attr"`
	Base    string   `xml:"base,attr"`
	Updir   *Updir   `xml:"updir"`
	Files   []File   `xml:"file"`
	Dirs    []Dir    `xml:"dir"`
}

type Updir struct {
	XMLName xml.Name `xml:"updir"`
	Href    string   `xml:"href,attr"`
}

type File struct {
	XMLName xml.Name `xml:"file"`
	Name    string   `xml:"name,attr"`
	Href    string   `xml:"href,attr"`
}

type Dir struct {
	XMLName xml.Name `xml:"dir"`
	Name    string   `xml:"name,attr"`
	Href    string   `xml:"href,attr"`
}

// Directory represents a language directory
type Directory struct {
	Code        string // Original directory name (e.g., "grekiska")
	Name        string // Translated name (e.g., "Greek")
	URL         string
	Description string
	Selected    bool
}

// DownloadResult stores information about a download operation
type DownloadResult struct {
	Directory  Directory
	Success    bool
	FileCount  int
	Error      error
	TotalBytes int64
}

// LanguageMap translates directory codes to human-readable names
var LanguageMap = map[string]string{
	"albanska":        "Albanian",
	"amhariska":       "Amharic",
	"arabiska":        "Arabic",
	"azerbajdzjanska": "Azerbaijani",
	"bosniska":        "Bosnian",
	"engelska":        "English",
	"finska":          "Finnish",
	"grekiska":        "Greek",
	"kroatiska":       "Croatian",
	"nordkurdiska":    "Northern Kurdish (Kurmanji)",
	"pashto":          "Pashto",
	"persiska":        "Persian (Farsi)",
	"ryska":           "Russian",
	"serbiska":        "Serbian",
	"somaliska":       "Somali",
	"spanska":         "Spanish",
	"svenska":         "Swedish",
	"sydkurdiska":     "Southern Kurdish (Sorani)",
	"tigrinska":       "Tigrinya",
	"turkiska":        "Turkish",
}
