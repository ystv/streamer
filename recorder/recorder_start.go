package main

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func (v *Views) start(transporter Transporter) error {
	array := strings.Split(transporter.Payload.(RecorderStart).PathOut, "/")
	valid := false
	var path string

	if len(array) == 1 {
		path = array[0]
		valid = true
	} else {
		for i := 0; i < len(array)-1; i++ {
			path += array[i] + "/"
		}
		err := os.MkdirAll(v.Config.RecordingLocation+path, 0777)
		if err != nil {
			return fmt.Errorf("error creating %s: %w", v.Config.RecordingLocation+path, err)
		}
		_, err1 := os.Stat(v.Config.RecordingLocation + path)
		if os.IsNotExist(err1) {
			return fmt.Errorf("unable to get path: %w", err1)
		}
		temp := array[len(array)-1]
		_, err2 := os.Stat(v.Config.RecordingLocation + path + "/" + temp)
		if os.IsNotExist(err2) {
			path += array[len(array)-1]
			valid = true
		} else {
			split := strings.Split(temp, ".")
			loop := true
			i := 0
			for loop {
				_, err3 := os.Stat(v.Config.RecordingLocation + path + "/" + split[0] + "_" + string(rune(i)) + ".mkv")
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
		return fmt.Errorf("invalid path")
	}

	streamIn := "rtmp://" + v.Config.StreamServer + transporter.Payload.(RecorderStart).StreamIn
	path = v.Config.RecordingLocation + path

	finish := make(chan bool)

	err := v.cache.Add(transporter.Unique+"Finish", finish, cache.NoExpiration)
	if err != nil {
		return err
	}

	go func() {
		var i uint64
		for {
			v.cache.Delete(transporter.Unique)
			switch {
			case <-finish:
				return
			default:
				c := exec.Command("ffmpeg", "-i", "\""+streamIn+"\"", "-c", "copy", "\""+path+"_"+strconv.FormatUint(i, 10)+".mkv\"", ">>", "\"/logs/"+transporter.Unique+".txt\"", "2>&1")
				err := v.cache.Add(transporter.Unique, c, cache.NoExpiration)
				if err != nil {
					log.Println(err)
					return
				}
				if err = c.Run(); err != nil {
					log.Println("could not run command: ", err)
				}
				time.Sleep(500 * time.Millisecond)
			}
			i++
		}
	}()

	go func() {
		for {
			switch {
			case <-finish:
				cmd, ok := v.cache.Get(transporter.Unique)
				if !ok {
					log.Println("unable to get cmd from cache")
				}
				c1 := cmd.(*exec.Cmd)
				err := c1.Process.Kill()
				if err != nil {
					log.Println(err)
				}
				break
			default:
				time.Sleep(1 * time.Second) // This is so it doesn't spam constantly and take the entire CPU up
				break
			}
		}
	}()

	return nil
}
