package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	issues := 0

	// Check 1: Version
	v := Version
	if v == "" {
		v = "dev"
	}
	fmt.Printf("  Version ... %s\n", v)

	// Check 2: Go toolchain
	if goPath, err := exec.LookPath("go"); err == nil {
		if out, err := exec.Command(goPath, "version").Output(); err == nil {
			fmt.Printf("  Go ... %s\n", strings.TrimSpace(string(out)))
		} else {
			fmt.Println("  Go ... found but could not determine version")
		}
	} else {
		fmt.Println("  Go ... not found (truthsayer scans Go source, but Go is not required)")
	}

	// Check 3: Config validity
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("  Config ... unable to determine working directory: %v\n", err)
		return 1
	}
	cfg, err := config.Load(dir, configPath)
	if err != nil {
		fmt.Printf("  Config ... FAIL: %v\n", err)
		issues++
	} else if configPath == "" {
		candidate := dir + "/.truthsayer.toml"
		if _, statErr := os.Stat(candidate); statErr != nil {
			fmt.Println("  Config ... no config file (using defaults)")
		} else {
			fmt.Println("  Config ... ok")
		}
	} else {
		fmt.Println("  Config ... ok")
	}

	// Check 4: Active rule count
	reg := rules.DefaultRegistry()
	if cfg != nil {
		for _, id := range cfg.Rules.Disable {
			reg.Disable(id)
		}
	}
	enabled := reg.EnabledRules()
	disabled := len(reg.AllRules()) - len(enabled)
	if disabled > 0 {
		fmt.Printf("  Rules ... %d enabled, %d disabled\n", len(enabled), disabled)
	} else {
		fmt.Printf("  Rules ... %d enabled\n", len(enabled))
	}

	// Check 5: Go files in directory
	goFiles := 0
	shFiles := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		base := filepath.Base(path)
		if info.IsDir() && (base == ".git" || base == "vendor" || base == "node_modules") {
			return filepath.SkipDir
		}
		switch filepath.Ext(path) {
		case ".go":
			goFiles++
		case ".sh", ".bash":
			shFiles++
		}
		return nil
	})
	fmt.Printf("  Files ... %d Go, %d bash files found\n", goFiles, shFiles)

	fmt.Println()
	if issues > 0 {
		fmt.Printf("  %d issue(s) found. Fix above and re-run.\n", issues)
		return 1
	}

	if goFiles == 0 && shFiles == 0 {
		fmt.Println("  No scannable files found in current directory.")
		fmt.Println("  Run 'truthsayer scan <path>' on a directory with Go or bash files.")
		return 0
	}

	fmt.Println("  Ready to scan. Quick start:")
	fmt.Println()
	fmt.Println("    truthsayer scan .           # Scan current directory")
	fmt.Println("    truthsayer scan . --format json  # JSON output for CI")
	fmt.Println("    truthsayer rules            # List all detection rules")
	fmt.Println("    truthsayer watch .          # Watch for changes")
	fmt.Println("    truthsayer hook install .   # Install git pre-commit hook")
	return 0
}
