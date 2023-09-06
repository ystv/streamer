package main

import (
	"fmt"
	"github.com/comtop1/gomux"
	"log"
	"os"
	"os/exec"
)

func stop(unique string) {
	gomux.KillSession("STREAM RECORDING - "+unique, os.Stdout)
	file := "logs/" + unique + ".txt"
	cmd := exec.Command("/bin/rm", file)
	_, err := cmd.Output()
	if err != nil {
		log.Fatalf("echo %+v", err)
	}
	fmt.Println("echo RECORDER STOPPED!")
}
