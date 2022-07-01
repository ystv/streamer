package main

import (
	"fmt"
	"github.com/wricardo/gomux"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if strings.Contains(os.Args[0], "/var/folders") || strings.Contains(os.Args[0], "/tmp/go") || strings.Contains(os.Args[0], "./forwarder_stop") {
		if len(os.Args) != 2 {
			fmt.Println("echo", os.Args)
			log.Fatalf("echo Arguments error")
		}
		for i := 0; i < len(os.Args)-1; i++ {
			os.Args[i] = os.Args[i+1]
		}
	} else {
		if len(os.Args) != 1 {
			fmt.Println("echo", os.Args)
			log.Fatalf("echo Arguments error")
		}
	}
	unique := os.Args[0]
	gomux.KillSession("STREAM FORWARDER - "+unique, os.Stdout)
	files, err := filepath.Glob("logs/" + unique + "_*")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			panic(err)
		}
	}
	fmt.Println("echo FORWARDER STOPPED!")
}
