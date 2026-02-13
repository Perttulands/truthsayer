package main

import (
	"os"

	"github.com/perttulands/truthsayer/internal/cli"
)

func main() {
	os.Exit(cli.Run())
}
