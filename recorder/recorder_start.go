package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) start(transporter commonTransporter.Transporter) error {
	array := strings.Split(transporter.Payload.(commonTransporter.RecorderStart).PathOut, "/")
	valid := false
	var path string

	tempBaseFileName := strings.Split(array[len(array)-1], ".")
	baseFileNameArray := tempBaseFileName[0 : len(tempBaseFileName)-1]
	var baseFileName string
	for _, s := range baseFileNameArray {
		baseFileName += s
	}

	if len(array) == 1 {
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
	}
	if !valid {
		return fmt.Errorf("invalid path")
	}

	streamIn := fmt.Sprintf("rtmp://%s%s", v.Config.StreamServer, transporter.Payload.(commonTransporter.RecorderStart).StreamIn)
	path = v.Config.RecordingLocation + path

	finish := make(chan bool)

	err := v.cache.Add(fmt.Sprintf("%s_%s", transporter.Unique, finishChannelNameAppend), finish, cache.NoExpiration)
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
				// Checking if file exists
				_, err = os.Stat(fmt.Sprintf("\"%s%s_%d.mkv", path, baseFileName, i))
				if err == nil {
					break
				}
				c := exec.Command("ffmpeg", "-i", fmt.Sprintf("\"%s\"", streamIn), "-c", "copy", fmt.Sprintf("\"%s%s_%d.mkv", path, baseFileName, i)+".mkv\"", ">>", "\"/logs/"+transporter.Unique+".txt\"", "2>&1")
				if err = v.cache.Add(transporter.Unique, c, cache.NoExpiration); err != nil {
					log.Println(err)
					return
				}
				err = c.Run()
				if err != nil {
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
				err = c1.Process.Kill()
				if err != nil {
					log.Println(err)
				}
				v.cache.Delete(transporter.Unique)
				return
			default:
				time.Sleep(1 * time.Second) // This is so it doesn't spam constantly and take the entire CPU up
			}
		}
	}()

	return nil
}
