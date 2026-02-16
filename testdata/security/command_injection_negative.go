package main

import "os/exec"

func run(userInput string) {
	cmd := exec.Command("ls", "--", userInput)
	_ = cmd
}
