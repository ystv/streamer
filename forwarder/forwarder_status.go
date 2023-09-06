package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

//go:embed forwarder_status.sh
var statusScript string

//func main() {
//	if strings.Contains(os.Args[0], "/var/folders") || strings.Contains(os.Args[0], "/tmp/go") || strings.Contains(os.Args[0], "forwarder_status") {
//		if len(os.Args) != 4 {
//			log.Fatalf("Arguments error")
//		}
//		for i := 0; i < len(os.Args)-1; i++ {
//			os.Args[i] = os.Args[i+1]
//		}
//	} else {
//		if len(os.Args) != 3 {
//			log.Fatalf("Arguments error")
//		}
//	}
//
//	website, _ := strconv.ParseBool(os.Args[0])
//	streams, _ := strconv.Atoi(os.Args[1])
//	unique := os.Args[2]
//
//	var start int
//
//	if website {
//		start = 0
//	} else {
//		start = 1
//	}
//
//	m := make(map[string]string)
//
//	for i := start; i <= streams; i++ {
//		out, err := exec.Command("./forwarder_status.sh", unique, strconv.Itoa(i)).Output()
//		if err != nil {
//			log.Fatal(err.Error())
//		} else {
//			if i == 0 {
//				m["website~"] = "~" + string(append(out, '\u0000'))
//			} else {
//				m[strconv.Itoa(i)+"~"] = "~" + string(append(out, '\u0000'))
//			}
//		}
//	}
//
//	fmt.Println(m)
//}

func status(unique string, website bool, streams int) {
	var start int

	if website {
		start = 0
	} else {
		start = 1
	}

	m := make(map[string]string)

	for i := start; i <= streams; i++ {
		c := exec.Command("bash", "-s", "-", unique, strconv.Itoa(i), "|", "bash")
		c.Stdin = strings.NewReader(statusScript)

		stderr, _ := c.StderrPipe()
		b, err := c.Output()
		if err != nil {
			log.Fatalf("echo %+v", err)
		}
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Fatalf("echo %s", scanner.Text())
		}

		if err != nil {
			log.Fatal(err.Error())
		} else {
			if i == 0 {
				m["website~"] = "~" + string(append(b, '\u0000'))
			} else {
				m[strconv.Itoa(i)+"~"] = "~" + string(append(b, '\u0000'))
			}
		}
	}

	fmt.Println(m)
}
