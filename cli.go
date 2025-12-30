package main

import (
	"fmt"
	"os"

	"github.com/chriso345/clifford"
)

// CLIArgs holds the parsed command-line arguments
type CLIArgs struct {
	clifford.Clifford `name:"greg"`
	clifford.Help
	clifford.Version `version:"0.1.0"`

	// Subcommands
	Menu struct {
		clifford.Subcommand `name:"menu"`
		clifford.Help
		clifford.Desc `desc:"Navigate predefined multi-level menu"`

		Start struct {
			Value             string
			clifford.Clifford `short:"s" long:"start" desc:"Initial submenu id"`
		}
		Prompt struct {
			Value             string
			clifford.Clifford `short:"p" long:"prompt" desc:"Prompt text"`
		}
		Out struct {
			Value             string
			clifford.Clifford `short:"o" long:"out" desc:"Write selection to file"`
		}
		Header struct {
			Value             string
			clifford.Clifford `long:"header" desc:"Header text"`
		}
		NoConfig struct {
			Value             bool
			clifford.Clifford `long:"no-config" desc:"Do not load config file"`
		}
		MaxItems struct {
			Value             int
			clifford.Clifford `short:"n" long:"max-items" desc:"Override max items (-1 for auto)"`
		}
		LogLevel struct {
			Value             string
			clifford.Clifford `long:"log-level" desc:"Set log level (debug|info|warn|error)"`
		}
		DryRun struct {
			Value             bool
			clifford.Clifford `long:"dry-run" desc:"Do not execute actions; print selection instead"`
		}
		Timeout struct {
			Value             int
			clifford.Clifford `long:"timeout" desc:"Auto-exit after N seconds of inactivity (0 disables)"`
		}
	}

	Dmenu struct {
		clifford.Subcommand `name:"dmenu"`
		clifford.Desc       `desc:"Filter piped input"`

		Prompt struct {
			Value             string
			clifford.Clifford `short:"p" long:"prompt" desc:"Prompt text"`
		}
		Out struct {
			Value             string
			clifford.Clifford `short:"o" long:"out" desc:"Write selection to file"`
		}
		MaxItems struct {
			Value             int
			clifford.Clifford `short:"n" long:"max-items" desc:"Override max items (-1 for auto)"`
		}
		LogLevel struct {
			Value             string
			clifford.Clifford `long:"log-level" desc:"Set log level (debug|info|warn|error)"`
		}
		DryRun struct {
			Value             bool
			clifford.Clifford `long:"dry-run" desc:"Do not execute actions; print selection instead"`
		}
		Timeout struct {
			Value             int
			clifford.Clifford `long:"timeout" desc:"Auto-exit after N seconds of inactivity (0 disables)"`
		}
	}

	Apps struct {
		clifford.Subcommand `name:"apps"`
		clifford.Desc       `desc:"List and launch .desktop applications"`

		DesktopDir struct {
			Value             string
			clifford.Clifford `long:"desktop-dir" desc:"Path to .desktop files"`
		}
		LogLevel struct {
			Value             string
			clifford.Clifford `long:"log-level" desc:"Set log level (debug|info|warn|error)"`
		}
		DryRun struct {
			Value             bool
			clifford.Clifford `long:"dry-run" desc:"Do not launch apps; print selection instead"`
		}
	}
}

// ParseArgs parses command-line flags using Clifford
func ParseArgs() *CLIArgs {
	args := &CLIArgs{}

	if err := clifford.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing arguments:", err)
		os.Exit(1)
	}

	return args
}
