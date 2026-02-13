package cli

import "fmt"

// Version is set at build time via -ldflags "-X github.com/perttulands/truthsayer/internal/cli.Version=X.Y.Z"
var Version string

func runVersion() int {
	v := Version
	if v == "" {
		v = "dev"
	}
	fmt.Printf("truthsayer version %s\n", v)
	return 0
}
