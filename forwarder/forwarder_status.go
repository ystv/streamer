package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) status(transporter commonTransporter.Transporter) (commonTransporter.ForwarderStatusResponse, error) {
	start := 1

	if transporter.Payload.(commonTransporter.ForwarderStatus).Website {
		start = 0
	}

	fStatusResponse := commonTransporter.ForwarderStatusResponse{}
	log.Println(10)

	for i := start; i <= transporter.Payload.(commonTransporter.ForwarderStatus).Streams; i++ {
		log.Println("i", i)
		c := exec.Command("tail", "-n", "19", fmt.Sprintf("\"logs/%s_%d.txt\"", transporter.Unique, i), "|", "sed", "-e", "\"s/\r$//\"")

		var stdout, stderr bytes.Buffer
		c.Stdout = &stdout
		c.Stderr = &stderr
		log.Println(11)

		var errOut string

		err := c.Run()
		if err != nil {
			errOut = fmt.Sprintf("could not run command: %+v", err)
		}

		log.Println(12)
		//stderr, err := c.StderrPipe()
		errOut += stderr.String()
		//scanner := bufio.NewScanner(stderr)
		//for scanner.Scan() {
		//	errOut += "\n" + scanner.Text()
		//}
		log.Println(13)

		if len(errOut) != 0 {
			return commonTransporter.ForwarderStatusResponse{}, fmt.Errorf(errOut)
		}

		log.Println(14)
		if i == 0 {
			fStatusResponse.Website = stdout.String()
		} else {
			fStatusResponse.Streams[uint64(i)] = stdout.String()
		}
		log.Println(15)
	}
	log.Println(16)
	log.Printf("%#v", fStatusResponse)

	return fStatusResponse, nil
}
