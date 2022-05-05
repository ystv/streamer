package main

import (
	"fmt"
	"github.com/wricardo/gomux"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if strings.Contains(os.Args[0], "/var/folders") || strings.Contains(os.Args[0], "/tmp/go") || strings.Contains(os.Args[0], "./recorder_stop") {
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
	gomux.KillSession("STREAM RECORDING - "+unique, os.Stdout)
	cmd := exec.Command("rm " + unique + "*.txt")
	err := cmd.Run()
	if err != nil {
		log.Fatal("echo " + err.Error())
	} else {
		fmt.Println("echo RECORDER STOPPED!")
	}
}
