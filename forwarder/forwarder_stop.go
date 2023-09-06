package main

import (
	"fmt"
	"github.com/ystv/streamer/gomux"
	"log"
	"os"
	"path/filepath"
)

func stop(unique string) {
	gomux.KillSession("STREAM FORWARDER - "+unique, os.Stdout)
	files, err := filepath.Glob("logs/" + unique + "_*")
	if err != nil {
		log.Fatalf("echo %+v", err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			log.Fatalf("echo %+v", err)
		}
	}
	fmt.Println("echo FORWARDER STOPPED!")
}
