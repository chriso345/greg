package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var pendingExec string
var pendingVisible bool

type model struct {
	// generic TUI fields
	allItems    []string
	filtered    []string
	cursor      int
	input       string
	width       int
	height      int
	windowStart int
	cursorStack  []int
	windowStartStack []int
	config      *Config

	mode       string
	prompt     string
	out        string
	mainHeader string
	helpText   string

	// persistent menu mode fields
	isMenuMode bool
	menuStack  [][]Menu
	current    []Menu
	labels     []string
}

func (m model) Init() tea.Cmd { return tea.EnterAltScreen }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch ev := msg.(type) {

	case tea.KeyMsg:
		key := ev.String()

		// ESC
		if key == "esc" {
			if m.isMenuMode && len(m.menuStack) > 0 {
				// go up one menu level
				m.current = m.menuStack[len(m.menuStack)-1]
				m.menuStack = m.menuStack[:len(m.menuStack)-1]
				m.updateMenuLabels()
				// restore cursor/windowStart if available
				if len(m.cursorStack) > 0 {
					restored := m.cursorStack[len(m.cursorStack)-1]
					m.cursorStack = m.cursorStack[:len(m.cursorStack)-1]
					if restored >= len(m.filtered) {
						if len(m.filtered) == 0 {
							m.cursor = -1
						} else {
							m.cursor = len(m.filtered) - 1
						}
					} else {
						m.cursor = restored
					}
				} else {
					m.cursor = 0
				}
				if len(m.windowStartStack) > 0 {
					restoredWS := m.windowStartStack[len(m.windowStartStack)-1]
					m.windowStartStack = m.windowStartStack[:len(m.windowStartStack)-1]
					if restoredWS >= len(m.filtered) {
						m.windowStart = 0
					} else {
						m.windowStart = restoredWS
					}
				} else {
					m.windowStart = 0
				}
				return m, nil
			}
			m.cursor = -1
			return m, tea.Quit
		}

		if key == "ctrl+c" {
			m.cursor = -1
			return m, tea.Quit
		}

		// ENTER
		if key == "enter" {
			if m.isMenuMode {
				if len(m.filtered) == 0 {
					return m, nil
				}
				selected := m.filtered[m.cursor]

				for _, item := range m.current {

					if item.Label != selected {
						continue
					}

					// SUBMENU
					if len(item.Items) > 0 {
						// save cursor/windowStart for restoration when returning
						m.cursorStack = append(m.cursorStack, m.cursor)
						m.windowStartStack = append(m.windowStartStack, m.windowStart)
						m.menuStack = append(m.menuStack, m.current)
						m.current = item.Items
						m.updateMenuLabels()
						m.cursor = 0
						m.windowStart = 0
						return m, nil
					}

					// GENERATOR
					if item.Generator != "" {
						gen, err := expandGenerator(item.Generator)
						if err == nil {
							// save cursor/windowStart for restoration when returning
							m.cursorStack = append(m.cursorStack, m.cursor)
							m.windowStartStack = append(m.windowStartStack, m.windowStart)
							m.menuStack = append(m.menuStack, m.current)
							m.current = gen
							m.updateMenuLabels()
							m.cursor = 0
							m.windowStart = 0
						}
						return m, nil
					}

					// EXEC
					if item.Exec != "" {
						pendingExec = item.Exec
						pendingVisible = item.Visible
						return m, tea.Quit
					}
				}
				return m, nil
			}

			// normal (dmenu/apps) mode
			return m, tea.Quit
		}

		// Navigation
		switch key {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.windowStart {
					m.windowStart--
				}
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				if m.cursor >= m.windowStart+m.config.MaxItems {
					m.windowStart++
				}
			}

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
				m.filterItems()
				m.windowStart = 0
			}

		default:
			if len(key) == 1 {
				m.input += key
				m.filterItems()
				m.windowStart = 0
			}
		}

	case tea.WindowSizeMsg:
		m.width = ev.Width
		m.height = ev.Height
	}

	return m, nil
}

func (m *model) filterItems() {
	var src []string
	if m.isMenuMode {
		src = m.labels
	} else {
		src = m.allItems
	}

	if m.input == "" {
		m.filtered = src
		m.cursor = 0
		m.windowStart = 0
		return
	}

	var f []string
	for _, item := range src {
		if strings.Contains(strings.ToLower(item), strings.ToLower(m.input)) {
			f = append(f, item)
		}
	}

	m.filtered = f
	if m.cursor >= len(f) {
		m.cursor = 0
	}
	if m.windowStart >= len(f) {
		m.windowStart = 0
	}
}

func (m model) View() string {
	cfg := m.config

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(cfg.Colors.Title))
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Colors.Prompt))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Colors.Item))
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color(cfg.Colors.Selected)).
		Bold(true)
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cfg.Colors.Help))

	header := ""
	if m.mainHeader != "" {
		header = titleStyle.Render(m.mainHeader) + helpStyle.Render(m.helpText) + "\n"
	}

	prompt := fmt.Sprintf("%s %s\n\n", promptStyle.Render(m.prompt), m.input)

	start := m.windowStart
	end := min(start+m.config.MaxItems, len(m.filtered))
	visible := m.filtered[start:end]

	list := ""
	for i, item := range visible {
		if start+i == m.cursor {
			list += selectedStyle.Render(" > "+item) + "\n"
		} else {
			list += itemStyle.Render("   "+item) + "\n"
		}
	}

	if len(m.filtered) == 0 {
		list += helpStyle.Render("   no matches found")
	}

	content := lipgloss.JoinVertical(lipgloss.Left, header, prompt, list)
	return lipgloss.NewStyle().Margin(1, 2).Render(content)
}

func RunTUIWithItems(cfg *Config, mode model, items []string, apps []AppEntry) (string, error) {
	p := tea.NewProgram(mode, tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	mod := m.(model)

	if len(mod.filtered) == 0 || mod.cursor == -1 {
		return "", nil
	}

	selected := mod.filtered[mod.cursor]

	switch mod.mode {
	case "dmenu":
		if mod.out != "" {
			if err := os.MkdirAll(filepath.Dir(mod.out), 0755); err != nil {
				return "", fmt.Errorf("failed to create output directory: %w", err)
			}
			if err := os.WriteFile(mod.out, []byte(selected+"\n"), 0644); err != nil {
				return "", fmt.Errorf("failed to write selection: %w", err)
			}
		} else {
			fmt.Println(selected)
		}

	case "apps":
		for _, app := range apps {
			if app.Name == selected {
				return "", launchDesktopFile(app.Path)
			}
		}

	case "menu":
		return selected, nil
	}

	return "", nil
}

func initialModelWithItems(cfg *Config, args *CLIArgs, items []string) model {
	prompt := args.Prompt.Value
	if prompt == "" {
		prompt = "search>"
	}

	header := args.Header.Value
	helpText := ""
	if header == "" {
		header = "greg"
		helpText = " - type to filter, ↑↓ to move, enter to select"
	}

	return model{
		allItems:    items,
		filtered:    items,
		config:      cfg,
		mode:        args.Mode.Value,
		prompt:      prompt,
		out:         args.Out.Value,
		mainHeader:  header,
		helpText:    helpText,
		windowStart: 0,
	}
}

func initialPersistentMenuModel(cfg *Config, _ *CLIArgs, menu *MenuConfig) model {
	prompt := menu.Prompt
	if prompt == "" {
		prompt = "search>"
	}

	header := menu.Title
	if header == "" {
		header = "greg"
	}

	m := model{
		config:     cfg,
		mode:       "menu",
		prompt:     prompt,
		mainHeader: header,
		helpText:   " - type to filter, ↑↓ to move, enter to select, esc to go back",

		isMenuMode: true,
		current:    menu.Menu,
		menuStack:  [][]Menu{},
	}

	m.updateMenuLabels()
	return m
}

// / -------------------------------------------------------------------------
func (m *model) updateMenuLabels() {
	m.input = ""
	m.labels = m.labels[:0]
	for _, item := range m.current {
		m.labels = append(m.labels, item.Label)
	}
	m.filtered = m.labels
}
