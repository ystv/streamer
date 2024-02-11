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
	fStatusResponse.Streams = make(map[uint64]string)
	log.Println(10)

	for i := start; i <= transporter.Payload.(commonTransporter.ForwarderStatus).Streams; i++ {
		log.Println("i", i)
		//c := exec.Command("tail", "-n", "19", fmt.Sprintf("\"logs/%s_%d.txt\"", transporter.Unique, i), "|", "sed", "-e", "\"s/\r$//\"")

		c1 := exec.Command("tail", "-n", "19", fmt.Sprintf("\"logs/%s_%d.txt\"", transporter.Unique, i))
		//c2 := exec.Command("sed", "-e", "'s/\r$//'")

		var stdout, stderr bytes.Buffer
		//c2.Stdout = &stdout
		//c2.Stderr = &stderr
		c1.Stdout = &stdout
		c1.Stderr = &stderr
		log.Println(11)

		//r, w := io.Pipe()
		//c1.Stdout = w
		//c2.Stdin = r

		log.Println(12)

		//_ = c1.Start()
		//_ = c2.Start()
		//_ = c1.Wait()
		//_ = w.Close()
		//_ = c2.Wait()

		var errOut string

		err := c1.Run()
		if err != nil {
			errOut = fmt.Sprintf("could not run command: %+v", err)
		}

		log.Println(13)
		//stderr, err := c.StderrPipe()
		errOut += stderr.String()
		//scanner := bufio.NewScanner(stderr)
		//for scanner.Scan() {
		//	errOut += "\n" + scanner.Text()
		//}
		log.Println(14)

		if len(errOut) != 0 {
			return commonTransporter.ForwarderStatusResponse{}, fmt.Errorf(errOut)
		}

		log.Println(15)
		if i == 0 {
			fStatusResponse.Website = stdout.String()
		} else {
			fStatusResponse.Streams[uint64(i)] = stdout.String()
		}
		log.Println(16)
	}
	log.Println(17)
	log.Printf("%#v", fStatusResponse)

	return fStatusResponse, nil
}
