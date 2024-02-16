package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) status(transporter commonTransporter.Transporter) (string, error) {
	c := exec.Command("tail", "-n", "19", fmt.Sprintf("\"logs/%s.txt\"", transporter.Unique), "|", "sed", "-e", "\"s/\r$//\"")

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

	var response string
	tempRespArr := strings.Split(strings.TrimRight(stdout.String(), "\r"), "\r")
	if len(tempRespArr) == 0 {
		response = "failed to get message response from recorder"
	} else {
		response = strings.ReplaceAll(tempRespArr[0], "\n", "<br>")
		response += "<br>"
		response += tempRespArr[len(tempRespArr)-1]
	}
	return response, nil
}
