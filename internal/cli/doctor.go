package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/perttulands/truthsayer/internal/config"
	"github.com/perttulands/truthsayer/internal/rules"
)

func runDoctor(args []string) int {
	configPath, remainingArgs := parseConfigFlag(args)
	if len(remainingArgs) > 0 {
		fmt.Fprintf(os.Stderr, "error: doctor does not accept positional arguments: %s\n", strings.Join(remainingArgs, " "))
		return 2
	}

	fmt.Println("truthsayer doctor")
	fmt.Println()

	// Check 1: Installation
	fmt.Println("  Installation ... ok")

	// Check 2: Config validity
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("  Config ... unable to determine working directory: %v\n", err)
		return 1
	}
	cfg, err := config.Load(dir, configPath)
	if err != nil {
		fmt.Printf("  Config ... config invalid: %v\n", err)
		return 1
	}

	if configPath == "" {
		candidate := dir + "/.truthsayer.toml"
		if _, statErr := os.Stat(candidate); statErr != nil {
			fmt.Println("  Config ... no config file (using defaults)")
		} else {
			fmt.Println("  Config ... config valid")
		}
	} else {
		fmt.Println("  Config ... config valid")
	}

	// Check 3: Active rule count
	reg := rules.DefaultRegistry()
	for _, id := range cfg.Rules.Disable {
		reg.Disable(id)
	}
	enabled := reg.EnabledRules()
	fmt.Printf("  Rules ... %d rules enabled\n", len(enabled))

	return 0
}
