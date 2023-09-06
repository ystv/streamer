package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

//go:embed recorder_status.sh
var statusScript string

func status(unique string) {
	c := exec.Command("bash", "-s", "-", unique, "|", "bash")

	c.Stdin = strings.NewReader(statusScript)

	stderr, _ := c.StderrPipe()
	b, err := c.Output()
	if err != nil {
		log.Fatalf("echo %+v", err)
	}
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		log.Fatalf("echo %s", scanner.Text())
	}
	fmt.Printf("echo %s", string(b))
}
