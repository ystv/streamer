package main

import (
	_ "embed"
	"fmt"
	"github.com/ystv/streamer/gomux"
	"os"
	"strconv"
)

//go:embed forwarder_start.sh
var startScript string

func start(unique, streamIn, websiteOut, streamServer string) {
	var serversKeys []string
	for i := 3; i < len(os.Args)-1; i++ {
		serversKeys = append(serversKeys, os.Args[i])
	}

	sessionName := "STREAM FORWARDER - " + unique

	s := gomux.NewSession(sessionName, os.Stdout)

	w1 := s.AddWindow("FORWARDING - 0")

	var panes []*gomux.Pane

	if websiteOut != "no" {
		panes = append(panes, w1.Pane(0))
		panes[0].Exec("./forwarder_start.sh " + streamServer + streamIn + " " + streamServer + "live/" + websiteOut + " " + unique + " " + strconv.Itoa(0) + " | bash")
	} else {
		panes = append(panes, w1.Pane(0))
		panes[0].Exec("echo No website stream")
	}
	j := 1
	k := 0
	for i := 0; i < len(serversKeys); i = i + 2 {
		if (i%8) == 0 && i != 0 {
			k++
			w1 = s.AddWindow("FORWARDING - " + strconv.Itoa(k))
			panes = append(panes, w1.Pane(0))
		}
		panes = append(panes, w1.Pane(0).Split())
		fmt.Println("echo", (i/2)+1)
		panes[(i/2)+1].Exec("./forwarder_start.sh " + streamServer + streamIn + " " + serversKeys[i] + serversKeys[i+1] + " " + unique + " " + strconv.Itoa(j) + " | bash")
		j++
	}

	fmt.Println("echo FORWARDER STARTED!")
}
