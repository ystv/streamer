package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) status(transporter commonTransporter.Transporter) (string, error) {
	c := exec.Command("tail", "-n", "26", fmt.Sprintf("\"logs/%s.txt\"", transporter.Unique))

	var stdout bytes.Buffer
	c.Stdout = &stdout

	var errOut string

	if err := c.Run(); err != nil {
		errOut = fmt.Sprintf("could not run command: %+v", err)
	}

	stderr, _ := c.StderrPipe()
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		errOut += "\n" + scanner.Text()
	}

	if len(errOut) != 0 {
		return "", fmt.Errorf(errOut)
	}

	return stdout.String(), nil
}
