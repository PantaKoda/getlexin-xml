package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"getlexin-xml/internal/models"
)

// For the TUI - this is a custom item type for our list
type item struct {
	Directory models.Directory
}

func (i item) Title() string {
	return i.Directory.Name
}

func (i item) Description() string {
	return fmt.Sprintf("(%s) - %s", i.Directory.Code, i.Directory.Description)
}

func (i item) FilterValue() string {
	return i.Directory.Name + " " + i.Directory.Code
}

// Keybinding mapping
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Toggle    key.Binding
	Help      key.Binding
	Quit      key.Binding
	Download  key.Binding
	SelectAll key.Binding
	None      key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Toggle, k.Download, k.SelectAll, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Toggle},
		{k.SelectAll, k.None, k.Download},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle selection"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Download: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "download selected"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "select/deselect all"),
	),
	None: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "select none"),
	),
}

// Model represents the TUI state
type Model struct {
	list          list.Model
	keys          keyMap
	help          help.Model
	directories   []models.Directory
	allSelected   bool
	baseURL       string
	outputDir     string
	concurrency   int
	quitting      bool
	ShowDownloads bool // Exported to be accessible from main
}

// Custom delegate for the list that shows checkboxes
type customItemDelegate struct {
	list.DefaultDelegate
}

// Make sure the delegate implements the ItemDelegate interface
func (d customItemDelegate) Height() int                               { return 2 }
func (d customItemDelegate) Spacing() int                              { return 1 }
func (d customItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d customItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	checked := "[ ]"
	if i.Directory.Selected {
		checked = "[✓]"
	}

	// Define some styles
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EE6FF8")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#777777"))

	// Determine if this item is selected
	isSelected := index == m.Index()

	// Create title string
	var title string
	if isSelected {
		title = selectedStyle.Render(fmt.Sprintf("%s %s", checked, i.Title()))
	} else {
		title = normalStyle.Render(fmt.Sprintf("%s %s", checked, i.Title()))
	}

	// Create description string
	desc := descStyle.Render(i.Description())

	// Set cursor indicator for selected item
	var cursor string
	if isSelected {
		cursor = selectedStyle.Render("> ")
	} else {
		cursor = "  "
	}

	// Print the item with cursor
	fmt.Fprintf(w, "%s%s\n  %s", cursor, title, desc)
}

// NewModel creates a new TUI model
func NewModel(directories []models.Directory, baseURL, outputDir string, concurrency int) Model {
	// Add an "All Languages" option at the top
	allOption := models.Directory{
		Code:        "all",
		Name:        "All Languages",
		Description: "Download all available language dictionaries",
		Selected:    false,
	}
	allDirs := append([]models.Directory{allOption}, directories...)

	// Create list items
	items := make([]list.Item, len(allDirs))
	for i, dir := range allDirs {
		items[i] = item{Directory: dir}
	}

	// Custom delegate with checkboxes
	delegate := customItemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}

	// Create the list with pagination
	const pageSize = 5
	l := list.New(items, delegate, pageSize, 15) // 15 is height, which will be adjusted by the terminal

	// Set styles for the list
	styles := list.DefaultStyles()
	styles.Title = styles.Title.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#0066CC")).
		Padding(0, 1)
	l.Styles = styles

	l.Title = "Available Lexin Language Dictionaries"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowPagination(true) // Show pagination indicator

	// Help model
	h := help.New()

	return Model{
		list:          l,
		keys:          keys,
		help:          h,
		directories:   allDirs,
		allSelected:   false,
		baseURL:       baseURL,
		outputDir:     outputDir,
		concurrency:   concurrency,
		quitting:      false,
		ShowDownloads: false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Toggle):
			// Toggle the selected state of the current item
			i := m.list.Index()
			if i >= 0 && i < len(m.directories) {
				// Special handling for "All Languages" option (index 0)
				if i == 0 {
					// Toggle all directories
					newState := !m.directories[0].Selected
					m.directories[0].Selected = newState
					m.allSelected = newState

					// If "All Languages" is selected, select all others
					for j := 1; j < len(m.directories); j++ {
						m.directories[j].Selected = newState
					}
				} else {
					// Toggle individual language
					m.directories[i].Selected = !m.directories[i].Selected

					// Update "All Languages" based on other selections
					allSelected := true
					for j := 1; j < len(m.directories); j++ {
						if !m.directories[j].Selected {
							allSelected = false
							break
						}
					}
					m.directories[0].Selected = allSelected
					m.allSelected = allSelected
				}

				// Update the list items
				items := make([]list.Item, len(m.directories))
				for i, dir := range m.directories {
					items[i] = item{Directory: dir}
				}
				m.list.SetItems(items)
			}

		case key.Matches(msg, m.keys.SelectAll):
			// Select or deselect all items
			newState := !m.allSelected
			m.allSelected = newState

			// Update all directories
			for i := range m.directories {
				m.directories[i].Selected = newState
			}

			// Update the list items
			items := make([]list.Item, len(m.directories))
			for i, dir := range m.directories {
				items[i] = item{Directory: dir}
			}
			m.list.SetItems(items)

		case key.Matches(msg, m.keys.None):
			// Deselect all
			m.allSelected = false

			for i := range m.directories {
				m.directories[i].Selected = false
			}

			// Update the list items
			items := make([]list.Item, len(m.directories))
			for i, dir := range m.directories {
				items[i] = item{Directory: dir}
			}
			m.list.SetItems(items)

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil

		case key.Matches(msg, m.keys.Download):
			// Start downloading selected directories
			var selectedDirs []models.Directory

			// Check if "All Languages" is selected
			if m.directories[0].Selected {
				// Skip the "All Languages" option itself
				selectedDirs = m.directories[1:]
			} else {
				// Get individually selected languages
				for i := 1; i < len(m.directories); i++ {
					if m.directories[i].Selected {
						selectedDirs = append(selectedDirs, m.directories[i])
					}
				}
			}

			if len(selectedDirs) > 0 {
				m.ShowDownloads = true
				// Return selected directories to start download
				return m, func() tea.Msg {
					return selectedDirs
				}
			}
		}

	case []models.Directory:
		// This is the return message from the download command
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return "Exiting..."
	}

	if m.ShowDownloads {
		return "Starting downloads...\n\nPress Ctrl+C to exit the program"
	}

	// Count selected items (excluding "All Languages" option)
	selectedCount := 0
	for i := 1; i < len(m.directories); i++ {
		if m.directories[i].Selected {
			selectedCount++
		}
	}

	var s strings.Builder

	s.WriteString("\n")
	s.WriteString(m.list.View())
	s.WriteString("\n\n")

	// Show selected count and help text without using SetStatusMessage
	statusStyle := lipgloss.NewStyle()
	if selectedCount > 0 {
		statusStyle = statusStyle.Foreground(lipgloss.Color("#10a010"))
	} else {
		statusStyle = statusStyle.Foreground(lipgloss.Color("#ffffff"))
	}

	statusText := fmt.Sprintf("Selected: %d languages • Use ↑/↓ to navigate • Space to toggle selection", selectedCount)
	s.WriteString(statusStyle.Render(statusText))
	s.WriteString("\n\n")

	// Show help
	helpView := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Render(m.help.View(m.keys))
	s.WriteString(helpView)

	return s.String()
}

// GetSelectedDirectories returns the directories selected by the user
func (m *Model) GetSelectedDirectories(originalDirs []models.Directory) []models.Directory {
	var selectedDirs []models.Directory

	// Check if "All Languages" is selected
	if m.directories[0].Selected {
		// "All Languages" is selected, download all (except the "All" item itself)
		return originalDirs
	}

	// Get individually selected languages
	for i := 1; i < len(m.directories); i++ {
		if m.directories[i].Selected {
			// Find the corresponding directory from the original list
			for _, d := range originalDirs {
				if d.Code == m.directories[i].Code {
					selectedDirs = append(selectedDirs, d)
					break
				}
			}
		}
	}

	return selectedDirs
}
