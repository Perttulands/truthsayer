package main

import "os/exec"

func run(userInput string) {
	cmd := exec.Command("sh", "-c", "ls "+userInput)
	_ = cmd
}
