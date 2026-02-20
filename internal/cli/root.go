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
	case "doctor":
		return runDoctor(os.Args[2:])
	case "judge":
		return runJudge(os.Args[2:])
	case "debt":
		return runDebt(os.Args[2:])
	case "senate":
		return runSenate(os.Args[2:])
	case "hook":
		if len(os.Args) > 2 && os.Args[2] == "install" {
			return runHookInstall(os.Args[3:])
		}
		return runHook(os.Args[2:])
	case "ci":
		if len(os.Args) > 2 && os.Args[2] == "init" {
			return runCIInit(os.Args[3:])
		}
		return runCI(os.Args[2:])
	case "--version", "-v", "version":
		return runVersion()
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
	fmt.Fprintln(os.Stderr, "truthsayer — development anti-pattern scanner for Go, JS/TS, Python, and bash")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage: truthsayer <command> [options]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  scan <path>          Scan a directory for anti-patterns")
	fmt.Fprintln(os.Stderr, "  check <file>         Scan a single file")
	fmt.Fprintln(os.Stderr, "  watch <path>         Watch for changes and scan modified files")
	fmt.Fprintln(os.Stderr, "  report <path>        Generate JSON report to file")
	fmt.Fprintln(os.Stderr, "  judge <findings.json>  Judge findings with precedents + LLM")
	fmt.Fprintln(os.Stderr, "  debt [path]          List advisory debt entries")
	fmt.Fprintln(os.Stderr, "  senate parse <file>  Parse and validate Senate verdict file")
	fmt.Fprintln(os.Stderr, "  senate apply <file> [repo]  Apply approved Senate amendments")
	fmt.Fprintln(os.Stderr, "  rules                List all detection rules")
	fmt.Fprintln(os.Stderr, "  rules --enabled      List only enabled rules (respects config)")
	fmt.Fprintln(os.Stderr, "  rules --lang <langs>  List rules for specific languages")
	fmt.Fprintln(os.Stderr, "  doctor               Check installation, config, and readiness")
	fmt.Fprintln(os.Stderr, "  hook <path>          Run as pre-commit hook (scan staged files)")
	fmt.Fprintln(os.Stderr, "  hook install <path>  Install git pre-commit hook")
	fmt.Fprintln(os.Stderr, "  ci [path]            CI gate — scan changed lines only")
	fmt.Fprintln(os.Stderr, "  ci init <path>       Generate GitHub Actions workflow")
	fmt.Fprintln(os.Stderr, "  --version            Print version")
	fmt.Fprintln(os.Stderr, "  --help               Show this help")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Scan options:")
	fmt.Fprintln(os.Stderr, "  --format json        Output JSON instead of terminal format")
	fmt.Fprintln(os.Stderr, "  --lang <langs>       Scan only specific languages (e.g., go,python,js)")
	fmt.Fprintln(os.Stderr, "  --parallel [n]       Scan files concurrently (default workers: NumCPU)")
	fmt.Fprintln(os.Stderr, "  --config <path>      Use a custom config file (default: .truthsayer.toml)")
	fmt.Fprintln(os.Stderr, "  --use-precedents     Suppress findings explicitly allowed in precedents.json")
	fmt.Fprintln(os.Stderr, "  --create-beads       Create problem beads for errors")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Judge options:")
	fmt.Fprintln(os.Stderr, "  --format json        Output verdicts as JSON (default)")
	fmt.Fprintln(os.Stderr, "  --precedents <path>  Precedent store path (default: <input-dir>/precedents.json)")
	fmt.Fprintln(os.Stderr, "  --debt <path>        Advisory debt path (default: <input-dir>/.truthsayer-debt.json)")
	fmt.Fprintln(os.Stderr, "  --law-candidates <path>  Consistent-ruling candidate log path")
	fmt.Fprintln(os.Stderr, "  --law-updates <path>     Markdown law update proposals path")
	fmt.Fprintln(os.Stderr, "  --law-threshold n    Candidate threshold for repeated same-pattern rulings (default: 10)")
	fmt.Fprintln(os.Stderr, "  --metrics <path>     Cost metrics JSONL path (default: <input-dir>/.truthsayer-cost.jsonl)")
	fmt.Fprintln(os.Stderr, "  --budget n           Spend cap in USD for LLM judgments (0 = unlimited)")
	fmt.Fprintln(os.Stderr, "  --min-confidence n   Minimum precedent confidence included in prompt context (0-1)")
	fmt.Fprintln(os.Stderr, "  --auto-apply-threshold n  Skip LLM when matching precedent confidence is above n (default: 0.9)")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Language aliases:")
	fmt.Fprintln(os.Stderr, "  go, js/javascript, ts/typescript, python/py, bash/shell/sh")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Exit codes:")
	fmt.Fprintln(os.Stderr, "  0  No error-severity findings")
	fmt.Fprintln(os.Stderr, "  1  Error-severity findings present")
	fmt.Fprintln(os.Stderr, "  2  Tool error (bad config, invalid path)")
}
