package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) status(transporter commonTransporter.Transporter) (string, error) {
	cmd := exec.Command("tail", "-n", "26", fmt.Sprintf("/logs/%s.txt", transporter.Unique))

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
		return "", fmt.Errorf(errOut)
	}

	var response string
	tempRespArr := strings.Split(strings.TrimRight(stdout.String(), "\r"), "\r")
	if len(tempRespArr) == 0 {
		response = "failed to get message response from recorder"
	} else {
		response = strings.ReplaceAll(tempRespArr[0], "\n", "<br>")
		response = strings.TrimSpace(response)
		response = strings.TrimRight(response, "size=       0kB time=00:00:00.00 bitrate=N/A speed=N/A")
		response = strings.TrimRight(response, "size=       0kB time=00:00:00.00 bitrate=N/A speed=   0x")
		response += tempRespArr[len(tempRespArr)-1]
	}
	return response, nil
}
