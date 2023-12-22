package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) status(transporter commonTransporter.Transporter) (commonTransporter.ForwarderStatusResponse, error) {
	var start int

	if transporter.Payload.(commonTransporter.ForwarderStatus).Website {
		start = 0
	} else {
		start = 1
	}

	fStatusResponse := commonTransporter.ForwarderStatusResponse{}

	for i := start; i <= transporter.Payload.(commonTransporter.ForwarderStatus).Streams; i++ {
		c := exec.Command("tail", "-n", "25", fmt.Sprintf("\"logs/%s_%d.txt\"", transporter.Unique, i))

		var stdout bytes.Buffer
		c.Stdout = &stdout

		var errOut string

		err := c.Run()
		if err != nil {
			errOut = fmt.Sprintf("could not run command: %+v", err)
		}

		stderr, _ := c.StderrPipe()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			errOut += "\n" + scanner.Text()
		}

		if len(errOut) != 0 {
			return commonTransporter.ForwarderStatusResponse{}, fmt.Errorf(errOut)
		}

		if i == 0 {
			fStatusResponse.Website = stdout.String()
		} else {
			fStatusResponse.Streams[uint64(i)] = stdout.String()
		}
	}

	return fStatusResponse, nil
}
