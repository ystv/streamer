package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) status(transporter commonTransporter.Transporter) (string, error) {
	logs := fmt.Sprintf("/logs/%s.txt", transporter.Unique)

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
		return "", errors.New(errOut)
	}

	var response string
	tempRespArr := strings.Split(strings.TrimRight(stdout.String(), "\r"), "\r")
	if len(tempRespArr) == 0 {
		response = "failed to get message response from recorder"
	} else {
		response = strings.ReplaceAll(tempRespArr[0], "\n", "<br>")
		response = strings.TrimSpace(response)
		baseTrim := "size=       0kB time=00:00:00.00 bitrate=N/A speed="
		response = strings.TrimRight(response, baseTrim+"N/A")
		response = strings.TrimRight(response, baseTrim+"   0x")
		response += tempRespArr[len(tempRespArr)-1]
	}
	return response, nil
}
