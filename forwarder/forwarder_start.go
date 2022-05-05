package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/wricardo/gomux"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	if strings.Contains(os.Args[0], "/var/folders") || strings.Contains(os.Args[0], "/tmp/go") || strings.Contains(os.Args[0], "./forwarder_start") {
		if len(os.Args) < 5 {
			log.Fatalf("echo Arguments error")
		}
		for i := 0; i < len(os.Args)-1; i++ {
			os.Args[i] = os.Args[i+1]
		}
	} else {
		if len(os.Args) < 4 {
			log.Fatalf("echo Arguments error")
		}
	}
	streamIn := os.Args[0]
	websiteOut := os.Args[1]
	unique := os.Args[2]
	var serversKeys []string
	for i := 3; i < len(os.Args)-1; i++ {
		serversKeys = append(serversKeys, os.Args[i])
	}

	sessionName := "STREAM FORWARDER - " + unique

	s := gomux.NewSession(sessionName, os.Stdout)

	w1 := s.AddWindow("FORWARDING")

	var panes []*gomux.Pane

	err := godotenv.Load()
	if err != nil {
		fmt.Printf("echo Error loading .env file: %s", err)
	} else {
		streamServer := os.Getenv("STREAM_SERVER")
		if websiteOut != "no" {
			panes = append(panes, w1.Pane(0))
			panes[0].Exec("./forwarder_script.sh " + streamServer + streamIn + " " + streamServer + "live/" + websiteOut + " " + unique + " " + strconv.Itoa(0) + " | bash")
			//panes[0].Exec("./forwarder_script.sh " + streamServer + streamIn + " " + "rtmp://stream.ystv.co.uk/live/" + websiteOut + " " + unique + " " + strconv.Itoa(0) + " | bash")
			//panes[0].Exec("ffmpeg -i \"" + streamServer + streamIn + "\" -c copy -f flv \"" + streamServer + "live/" + websiteOut + "\"")
		} else {
			panes = append(panes, w1.Pane(0))
			panes[0].Exec("echo No website stream")
		}
		j := 1
		for i := 0; i < len(serversKeys); i = i + 2 {
			panes = append(panes, w1.Pane(0).Split())
			panes[(i/2)+1].Exec("./forwarder_script.sh " + streamServer + streamIn + " " + serversKeys[i] + serversKeys[i+1] + " " + unique + " " + strconv.Itoa(j) + " | bash")
			//panes[(i/2)+1].Exec("ffmpeg -i \"" + streamServer + streamIn + "\" -c copy -f flv \"" + serversKeys[i] + serversKeys[i+1] + "\"")
			j++
		}

		fmt.Println("echo FORWARDER STARTED!")
	}
}
