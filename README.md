# Lexin Dictionary Downloader

A Go application for downloading Lexin dictionary XML files. This tool provides a terminal user interface (TUI) to select and download Swedish bilingual dictionaries from the Lexin project.

## Features

- Interactive terminal UI for selecting languages
- Concurrent downloads with configurable concurrency
- Progress reporting during downloads
- Human-readable file sizes
- Organized output with metadata

## Installation

1. Ensure you have Go installed (version 1.18 or higher)
2. Clone this repository
3. Build the application:

```bash
go build -o lexin-downloader ./cmd/lexin
```

Or simply use the provided Makefile:

```bash
make build
```

## Usage

```bash
./lexin-downloader [options]
```

### Options

- `-out string`: Output directory for downloads (default "lexin_downloads")
- `-concurrency int`: Number of concurrent downloads (default 3)

### UI Controls

- **↑/↓ or j/k**: Navigate the list
- **Space**: Toggle selection of the current item
- **a**: Select/deselect all
- **n**: Deselect all
- **Enter**: Start downloading selected languages
- **?**: Toggle help view
- **q/ESC/Ctrl+C**: Quit

## Project Structure

```
lexin-downloader/
├── cmd/
│   └── lexin/
│       └── main.go       # Main entry point
├── internal/
│   ├── models/
│   │   └── types.go      # Data structures
│   ├── fetcher/
│   │   └── fetcher.go    # Handles XML fetching
│   ├── parser/
│   │   └── parser.go     # XML parsing
│   └── ui/
│       └── tui.go        # Terminal UI
├── go.mod
└── go.sum
```

## Output

Downloads are organized by language code in the specified output directory. Each language directory contains:

- The XML dictionary files
- An index.html file
- A metadata.txt file with download information

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal applications
- [go-humanize](https://github.com/dustin/go-humanize) - Human-readable formatting

## Development

To modify or extend this application:

1. Clone the repository
2. Install dependencies: `make deps`
3. Make your changes
4. Format the code: `make fmt`
5. Build and test: `make run`

## License

MIT