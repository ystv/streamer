package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) status(transporter commonTransporter.Transporter) (commonTransporter.ForwarderStatusResponse, error) {
	start := 1

	if transporter.Payload.(commonTransporter.ForwarderStatus).Website {
		start = 0
	}

	fStatusResponse := commonTransporter.ForwarderStatusResponse{}
	fStatusResponse.Streams = make(map[string]string)

	for i := start; i <= transporter.Payload.(commonTransporter.ForwarderStatus).Streams; i++ {
		cmd := exec.Command("tail", "-n", "26", fmt.Sprintf("/logs/%s_%d.txt", transporter.Unique, i))

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		var errOut string

		err := cmd.Run()
		if err != nil {
			errOut = fmt.Sprintf("could not run command: %+v", err)
		}

		errOut += stderr.String()

		if len(errOut) != 0 {
			return commonTransporter.ForwarderStatusResponse{}, fmt.Errorf(errOut)
		}

		var response string
		tempRespArr := strings.Split(strings.TrimRight(stdout.String(), "\r"), "\r")
		if len(tempRespArr) == 0 {
			response = fmt.Sprintf("failed to get message response from forwarder for stream %d", i)
		} else {
			response = strings.ReplaceAll(tempRespArr[0], "\n", "<br>")
			response = strings.TrimSpace(response)
			response = strings.TrimRight(response, "size=       0kB time=00:00:00.00 bitrate=N/A speed=N/A")
			response = strings.TrimRight(response, "size=       0kB time=00:00:00.00 bitrate=N/A speed=   0x")
			response += tempRespArr[len(tempRespArr)-1]
		}
		if i == 0 {
			fStatusResponse.Website = response
		} else {
			fStatusResponse.Streams[fmt.Sprintf("%d", i)] = response
		}
	}

	return fStatusResponse, nil
}
