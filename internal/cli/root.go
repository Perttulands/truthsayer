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
	fmt.Fprintln(os.Stderr, "  --help         Show this help")
}
