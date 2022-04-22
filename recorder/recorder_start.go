package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/wricardo/gomux"
	"log"
	"os"
	"strings"
)

func main() {
	fmt.Println("echo", os.Args)
	if strings.Contains(os.Args[0], "/var/folders") || strings.Contains(os.Args[0], "/tmp/go") || strings.Contains(os.Args[0], "./recorder_start") {
		if len(os.Args) != 4 {
			fmt.Println("echo " + string(rune(len(os.Args))))
			log.Fatalf("echo Arguments error")
		}
		for i := 0; i < len(os.Args)-1; i++ {
			os.Args[i] = os.Args[i+1]
		}
	} else {
		if len(os.Args) != 3 {
			fmt.Println("echo " + string(rune(len(os.Args))))
			log.Fatalf("echo Arguments error")
		}
	}
	streamIn := os.Args[0]
	pathOut := os.Args[1]
	unique := os.Args[2]
	array := strings.Split(pathOut, "/")
	valid := false
	var path string
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("echo Error loading .env file: %s", err)
	} else {
		pendingEdits := os.Getenv("PENDING_EDITS")
		if len(array) == 1 {
			path = array[0]
			valid = true
		} else {
			for i := 0; i < len(array)-1; i++ {
				path += array[i] + "/"
			}
			err = os.MkdirAll(pendingEdits+path, 0777)
			if err != nil {
				fmt.Println("echo " + path)
				fmt.Println("echo " + err.Error())
				log.Fatal("echo Error creating " + pendingEdits + path)
			}
			_, err1 := os.Stat(pendingEdits + path)
			if os.IsNotExist(err1) {
				fmt.Println(" echo RECORDER UNSUCCESSFUL!")
			} else {
				temp := array[len(array)-1]
				_, err2 := os.Stat(pendingEdits + path + "/" + temp)
				if os.IsNotExist(err2) {
					path += array[len(array)-1]
					valid = true
				} else {
					split := strings.Split(temp, ".")
					loop := true
					i := 0
					for loop {
						_, err3 := os.Stat(pendingEdits + path + "/" + split[0] + "_" + string(rune(i)) + ".mkv")
						if os.IsNotExist(err3) {
							path += split[0] + string(rune(i)) + ".mkv"
							loop = false
							valid = true
						} else {
							i++
						}
					}
				}
			}
		}
		if valid {
			streamServer := os.Getenv("STREAM_SERVER")

			sessionName := "STREAM RECORDING - " + unique

			s := gomux.NewSession(sessionName, os.Stdout)

			w1 := s.AddWindow("RECORDING")

			w1p0 := w1.Pane(0)
			w1p0.Exec("ffmpeg -i \"" + streamServer + streamIn + "\" -c copy -f mp4 \"" + pendingEdits + path + "\"")

			fmt.Println("echo RECORDER STARTED!")
		} else {
			log.Fatal("echo Invalid string")
		}
	}
}
