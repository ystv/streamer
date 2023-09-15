package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
)

func (v *Views) status(transporter Transporter) (ForwarderStatusResponse, error) {
	var start int

	if transporter.Payload.(ForwarderStatus).Website {
		start = 0
	} else {
		start = 1
	}

	fStatusResponse := ForwarderStatusResponse{}

	for i := start; i <= transporter.Payload.(ForwarderStatus).Streams; i++ {
		c := exec.Command("tail", "-n", "25", fmt.Sprintf("\"logs/%s_%d.txt\"", transporter.Unique, i))

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
			return ForwarderStatusResponse{}, fmt.Errorf(errOut)
		}

		if i == 0 {
			fStatusResponse.Website = stdout.String()
		} else {
			fStatusResponse.Streams[uint64(i)] = stdout.String()
		}
	}

	return fStatusResponse, nil
}
