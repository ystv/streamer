package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

//go:embed forwarder_status.sh
var statusScript string

func status(unique string, website bool, streams int) {
	var start int

	if website {
		start = 0
	} else {
		start = 1
	}

	m := make(map[string]string)

	for i := start; i <= streams; i++ {
		c := exec.Command("bash", "-s", "-", unique, strconv.Itoa(i), "|", "bash")
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

		if err != nil {
			log.Fatal(err.Error())
		} else {
			if i == 0 {
				m["website~"] = "~" + string(append(b, '\u0000'))
			} else {
				m[strconv.Itoa(i)+"~"] = "~" + string(append(b, '\u0000'))
			}
		}
	}

	fmt.Println(m)
}
