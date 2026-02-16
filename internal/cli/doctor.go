package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"

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

	// Check 4: Active rule count with per-language breakdown
	reg := rules.DefaultRegistry()
	if cfg != nil {
		for _, id := range cfg.Rules.Disable {
			reg.Disable(id)
		}
	}
	enabled := reg.EnabledRules()
	disabled := len(reg.AllRules()) - len(enabled)
	goRules, jstsRules, pyRules, bashRules := countRulesByLang(enabled)
	if disabled > 0 {
		fmt.Printf("  Rules ... %d enabled, %d disabled (%d Go, %d JS/TS, %d Python, %d bash)\n",
			len(enabled), disabled, goRules, jstsRules, pyRules, bashRules)
	} else {
		fmt.Printf("  Rules ... %d enabled (%d Go, %d JS/TS, %d Python, %d bash)\n",
			len(enabled), goRules, jstsRules, pyRules, bashRules)
	}

	// Check 5: Parser status
	if ok := checkParser(javascript.GetLanguage(), "var x = 1;"); ok {
		fmt.Println("  JS/TS AST parser ... available (tree-sitter)")
	} else {
		fmt.Println("  JS/TS AST parser ... unavailable")
		issues++
	}
	if ok := checkParser(python.GetLanguage(), "x = 1"); ok {
		fmt.Println("  Python AST parser ... available (tree-sitter)")
	} else {
		fmt.Println("  Python AST parser ... unavailable")
		issues++
	}

	// Check 6: Files in directory
	goFiles, jstsFiles, pyFiles, shFiles := 0, 0, 0, 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		base := filepath.Base(path)
		if info.IsDir() && (base == ".git" || base == "vendor" || base == "node_modules" || base == "__pycache__" || base == ".venv") {
			return filepath.SkipDir
		}
		switch filepath.Ext(path) {
		case ".go":
			goFiles++
		case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
			jstsFiles++
		case ".py", ".pyi":
			pyFiles++
		case ".sh", ".bash":
			shFiles++
		}
		return nil
	})
	fmt.Printf("  Files ... %d Go, %d JS/TS, %d Python, %d bash files found\n", goFiles, jstsFiles, pyFiles, shFiles)

	fmt.Println()
	if issues > 0 {
		fmt.Printf("  %d issue(s) found. Fix above and re-run.\n", issues)
		return 1
	}

	if goFiles == 0 && jstsFiles == 0 && pyFiles == 0 && shFiles == 0 {
		fmt.Println("  No scannable files found in current directory.")
		fmt.Println("  Run 'truthsayer scan <path>' on a directory with Go, JS/TS, Python, or bash files.")
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

// countRulesByLang counts enabled rules per language group.
func countRulesByLang(enabled []rules.Rule) (goCount, jstsCount, pyCount, bashCount int) {
	for _, r := range enabled {
		lang := classifyRule(r)
		switch lang {
		case "go":
			goCount++
		case "jsts":
			jstsCount++
		case "python":
			pyCount++
		case "bash":
			bashCount++
		}
	}
	return
}

// classifyRule determines a rule's primary language from its FileTypes.
func classifyRule(r rules.Rule) string {
	for _, ft := range r.FileTypes {
		switch ft {
		case ".go":
			return "go"
		case ".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs":
			return "jsts"
		case ".py", ".pyi":
			return "python"
		case ".sh", ".bash":
			return "bash"
		}
	}
	return ""
}

// checkParser verifies a tree-sitter parser can initialize and parse a snippet.
func checkParser(lang *sitter.Language, snippet string) bool {
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, []byte(snippet))
	if err != nil {
		return false
	}
	return tree.RootNode() != nil
}
