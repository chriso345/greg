package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

type AppEntry struct {
	Name string
	Path string
}

func main() {
	args := ParseArgs()

	if os.Getenv("GREG_DEBUG_ARGS") == "1" {
		fmt.Fprintf(os.Stderr, "DEBUG ARGS: %+v\n", args)
		os.Exit(1)
	}

	// Determine active subcommand
	modeName := ""
	if args.Menu.Subcommand {
		modeName = "menu"
	} else if args.Dmenu.Subcommand {
		modeName = "dmenu"
	} else if args.Apps.Subcommand {
		modeName = "apps"
	} else {
		modeName = "apps"
	}

	// Load config unless menu --no-config was requested
	var cfg *Config
	var err error
	if modeName == "menu" && args.Menu.NoConfig.Value {
		cfg = defaultConfig()
	} else {
		cfg, err = LoadConfig()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: using default config -", err)
			cfg = defaultConfig()
		}
	}

	// Apply per-subcommand flags to config
	switch modeName {
	case "menu":
		if args.Menu.MaxItems.Value != 0 {
			cfg.MaxItems = args.Menu.MaxItems.Value
		}
		if args.Menu.LogLevel.Value != "" {
			lvl := args.Menu.LogLevel.Value
			cfg.Log = !(lvl == "error" || lvl == "warn")
		}
	case "dmenu":
		if args.Dmenu.MaxItems.Value != 0 {
			cfg.MaxItems = args.Dmenu.MaxItems.Value
		}
		if args.Dmenu.LogLevel.Value != "" {
			lvl := args.Dmenu.LogLevel.Value
			cfg.Log = !(lvl == "error" || lvl == "warn")
		}
	case "apps":
		if args.Apps.LogLevel.Value != "" {
			lvl := args.Apps.LogLevel.Value
			cfg.Log = !(lvl == "error" || lvl == "warn")
		}
	}

	var items []string
	var appEntries []AppEntry

	switch modeName {
	case "dmenu":
		// Ensure piped input
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			fmt.Fprintln(os.Stderr, "Error: expected piped input, e.g., `ls | greg dmenu`.")
			os.Exit(1)
		}

		// Read piped items
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			items = append(items, scanner.Text())
		}

	case "menu":
		cfg.MaxItems = getMaxItems(cfg)

		mnu, err := loadMenu()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if args.Menu.Start.Value != "" {
			subMenu := findMenuByID(mnu.Menu, args.Menu.Start.Value)
			if subMenu == nil {
				// Shouldn't be fatal
				fmt.Fprintf(os.Stderr, "Warning: submenu with ID '%s' not found. Starting from root menu.\n", args.Menu.Start.Value)
			} else {
				mnu.Menu = subMenu.Items
			}
		}

		runMenu(mnu, cfg, args)
		os.Exit(0)

	case "apps":
		// apps: load .desktop files, allow override via flag
		dir := "/home/chris/.local/share/applications"
		if args.Apps.DesktopDir.Value != "" {
			dir = args.Apps.DesktopDir.Value
		}
		appEntries, err = readDesktopFiles(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading .desktop files:", err)
			os.Exit(1)
		}

		for _, app := range appEntries {
			items = append(items, app.Name)

			// Print debug info
			if cfg.Log {
				fmt.Printf("[DEBUG] Loaded app: %s (%s)\n", app.Name, app.Path)
			}
		}

	default:
		fmt.Fprintln(os.Stderr, "Error: unknown mode. Supported modes: dmenu, menu, apps")
		os.Exit(1)
	}

	cfg.MaxItems = getMaxItems(cfg)

	if cfg.Log {
		fmt.Printf("[DEBUG] Total apps loaded: %d\n", len(appEntries))
	}

	// Determine prompt/out/header to pass into TUI
	var finalPrompt, finalOut, finalHeader string
	switch modeName {
	case "menu":
		finalPrompt = args.Menu.Prompt.Value
		finalOut = args.Menu.Out.Value
		finalHeader = args.Menu.Header.Value
	case "dmenu":
		finalPrompt = args.Dmenu.Prompt.Value
		finalOut = args.Dmenu.Out.Value
		finalHeader = ""
	default: // apps
		finalPrompt = ""
		finalOut = ""
		finalHeader = ""
	}

	mode := initialModelWithItems(cfg, modeName, finalPrompt, finalOut, finalHeader, items)
	// set timeout and dry-run from CLI flags per subcommand
	switch modeName {
	case "menu":
		mode.timeout = args.Menu.Timeout.Value
		mode.dryRun = args.Menu.DryRun.Value
	case "dmenu":
		mode.timeout = args.Dmenu.Timeout.Value
		mode.dryRun = args.Dmenu.DryRun.Value
	case "apps":
		// apps has no timeout flag; keep default 0
		mode.dryRun = args.Apps.DryRun.Value
	}
	if _, err := RunTUIWithItems(cfg, mode, items, appEntries); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// readDesktopFiles returns the "Name=" entries from all .desktop files in the folder
func readDesktopFiles(dir string) ([]AppEntry, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var apps []AppEntry
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".desktop") {
			continue
		}

		path := filepath.Join(dir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var name string
		for line := range strings.SplitSeq(string(data), "\n") {
			if after, ok := strings.CutPrefix(line, "Name="); ok {
				name = after
				name = strings.TrimSpace(name)
				break
			}
		}
		if name != "" {
			apps = append(apps, AppEntry{Name: name, Path: path})
		}
	}

	return apps, nil
}

// getMaxItems calculates the number of visible items for the TUI.
// If cfg.MaxItems >= 0, it returns cfg.MaxItems.
// If cfg.MaxItems == -1, it auto-detects terminal height.
func getMaxItems(cfg *Config) int {
	if cfg.MaxItems > 0 {
		return cfg.MaxItems
	}

	height, _, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil || height < 5 {
		// Fallback if detection fails
		if cfg.Log {
			fmt.Printf("[DEBUG] Failed to get terminal size, using default max items 10: %v\n", err)
		}
		return cfg.DefaultMaxItems
	}

	// Reserve lines for header, prompt, margins, etc.
	reservedLines := 4
	maxItems := max(height-reservedLines, 1)
	return maxItems
}
