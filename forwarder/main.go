package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("echo", os.Args)
	if strings.Contains(os.Args[0], "/var/folders") || strings.Contains(os.Args[0], "/tmp/go") || strings.Contains(os.Args[0], "./recorder_start") {
		if len(os.Args) != 6 && len(os.Args) != 5 && len(os.Args) != 3 {
			fmt.Println("echo " + string(rune(len(os.Args))))
			log.Fatalf("echo Arguments error")
		}
		for i := 0; i < len(os.Args)-1; i++ {
			os.Args[i] = os.Args[i+1]
		}
	} else {
		if len(os.Args) != 5 && len(os.Args) != 4 && len(os.Args) != 2 {
			fmt.Println("echo " + string(rune(len(os.Args))))
			log.Fatalf("echo Arguments error")
		}
	}
	method := os.Args[0]
	switch method {
	case "start":
		start(os.Args[1], os.Args[2], os.Args[3], os.Args[4])
		break
	case "stop":
		stop(os.Args[1])
		break
	case "status":
		website, err := strconv.ParseBool(os.Args[2])
		if err != nil {
			log.Fatalf("echo unable to parse bool: %+v", err)
		}
		streams, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatalf("echo unable to parse int: %+v", err)
		}
		status(os.Args[1], website, streams)
		break
	default:
		log.Fatalf("echo Invalid method used: %s", method)
	}
}
