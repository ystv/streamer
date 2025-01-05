package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
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

	baseTrim := "size=       0kB time=00:00:00.00 bitrate=N/A speed="

	for i := start; i <= transporter.Payload.(commonTransporter.ForwarderStatus).Streams; i++ {
		logs := fmt.Sprintf("/logs/%s_%d.txt", transporter.Unique, i)

		cmd := exec.Command("tail", "-n", "26", logs)

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
			return commonTransporter.ForwarderStatusResponse{}, errors.New(errOut)
		}

		var response string
		tempRespArr := strings.Split(strings.TrimRight(stdout.String(), "\r"), "\r")
		if len(tempRespArr) == 0 {
			response = fmt.Sprintf("failed to get message response from forwarder for stream %d", i)
		} else {
			response = strings.ReplaceAll(tempRespArr[0], "\n", "<br>")
			response = strings.TrimSpace(response)
			response = strings.TrimRight(response, baseTrim+"N/A")
			response = strings.TrimRight(response, baseTrim+"   0x")
			response += tempRespArr[len(tempRespArr)-1]
		}
		if i == 0 {
			fStatusResponse.Website = response
		} else {
			fStatusResponse.Streams[strconv.Itoa(i)] = response
		}
	}

	return fStatusResponse, nil
}
