package main

import (
	_ "embed"
	"fmt"
	"github.com/wricardo/gomux"
	"log"
	"os"
	"os/exec"
	"strings"
)

//go:embed recorder_start.sh
var startScript string

func start(unique, streamIn, pathOut, recordingLocation, streamServer string) {
	array := strings.Split(pathOut, "/")
	valid := false
	var path string

	if len(array) == 1 {
		path = array[0]
		valid = true
	} else {
		for i := 0; i < len(array)-1; i++ {
			path += array[i] + "/"
		}
		err := os.MkdirAll(recordingLocation+path, 0777)
		if err != nil {
			fmt.Println("echo " + path)
			fmt.Println("echo " + err.Error())
			log.Fatal("echo Error creating " + recordingLocation + path)
		}
		_, err1 := os.Stat(recordingLocation + path)
		if os.IsNotExist(err1) {
			log.Fatal(" echo RECORDER UNSUCCESSFUL!")
		}
		temp := array[len(array)-1]
		_, err2 := os.Stat(recordingLocation + path + "/" + temp)
		if os.IsNotExist(err2) {
			path += array[len(array)-1]
			valid = true
		} else {
			split := strings.Split(temp, ".")
			loop := true
			i := 0
			for loop {
				_, err3 := os.Stat(recordingLocation + path + "/" + split[0] + "_" + string(rune(i)) + ".mkv")
				if os.IsNotExist(err3) {
					path += split[0] + string(rune(i)) + ".mkv"
					loop = false
					valid = true
					break
				}
				i++
			}
		}
	}
	if !valid {
		log.Fatal("echo Invalid string")
	}
	sessionName := "STREAM RECORDING - " + unique

	s := gomux.NewSession(sessionName, os.Stdout)

	w1 := s.AddWindow("RECORDING")

	w1p0 := w1.Pane(0)

	path = strings.ReplaceAll(path, ".mkv", "")

	c := exec.Command("sh", "-s", "-", streamServer+streamIn, fmt.Sprintf("\"%s%s\"", recordingLocation, path), unique, "|", "bash")

	c.Stdin = strings.NewReader(startScript)

	b, e := c.Output()
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(string(b))

	w1p0.Exec("./recorder_start.sh " + streamServer + streamIn + " \"" + recordingLocation + path + "\" " + unique + " | bash")

	fmt.Println("echo RECORDER STARTED!")

}
