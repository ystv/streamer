package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	if strings.Contains(os.Args[0], "/var/folders") || strings.Contains(os.Args[0], "/tmp/go") || strings.Contains(os.Args[0], "forwarder_status") {
		if len(os.Args) != 4 {
			log.Fatalf("Arguments error")
		}
		for i := 0; i < len(os.Args)-1; i++ {
			os.Args[i] = os.Args[i+1]
		}
	} else {
		if len(os.Args) != 3 {
			log.Fatalf("Arguments error")
		}
	}

	website, _ := strconv.ParseBool(os.Args[0])
	streams, _ := strconv.Atoi(os.Args[1])
	unique := os.Args[2]

	var start int

	if website {
		start = 0
	} else {
		start = 1
	}

	m := make(map[string]string)

	for i := start; i <= streams; i++ {
		out, err := exec.Command("./forwarder_status.sh", unique, strconv.Itoa(i)).Output()
		if err != nil {
			log.Fatal(err.Error())
		} else {
			if i == 0 {
				m["website~"] = "~" + string(append(out, '\u0000'))
			} else {
				m[strconv.Itoa(i)+"~"] = "~" + string(append(out, '\u0000'))
			}
		}
	}

	fmt.Println(m)
}
