package cli

import (
	"fmt"
	"os"
)

// Run is the main CLI entrypoint.
func Run() int {
	if len(os.Args) < 2 {
		printUsage()
		return 2
	}

	switch os.Args[1] {
	case "scan":
		return runScan(os.Args[2:])
	case "check":
		return runCheck(os.Args[2:])
	case "watch":
		return runWatch(os.Args[2:])
	case "report":
		return runReport(os.Args[2:])
	case "rules":
		return runRules(os.Args[2:])
	case "hook":
		if len(os.Args) > 2 && os.Args[2] == "install" {
			return runHookInstall(os.Args[3:])
		}
		return runHook(os.Args[2:])
	case "ci":
		if len(os.Args) > 2 && os.Args[2] == "init" {
			return runCIInit(os.Args[3:])
		}
		fmt.Fprintln(os.Stderr, "usage: truthsayer ci init <path>")
		return 2
	case "--help", "-h", "help":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		return 2
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: truthsayer <command> [options]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  scan <path>    Scan a directory for anti-patterns")
	fmt.Fprintln(os.Stderr, "  check <file>   Scan a single file for anti-patterns")
	fmt.Fprintln(os.Stderr, "  watch <path>   Watch a directory for changes and scan")
	fmt.Fprintln(os.Stderr, "  report <path>  Generate JSON report to file")
	fmt.Fprintln(os.Stderr, "  rules          List all available detection rules")
	fmt.Fprintln(os.Stderr, "  rules --enabled List only currently enabled rules")
	fmt.Fprintln(os.Stderr, "  hook <path>    Run pre-commit hook (scan staged files)")
	fmt.Fprintln(os.Stderr, "  hook install <path>  Install git pre-commit hook")
	fmt.Fprintln(os.Stderr, "  ci init <path>       Generate GitHub Actions workflow")
	fmt.Fprintln(os.Stderr, "  --help         Show this help")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Global options:")
	fmt.Fprintln(os.Stderr, "  --config <path>  Use a custom config file (default: .truthsayer.toml)")
}
