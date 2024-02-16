package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	ffmpeg "github.com/u2takey/ffmpeg-go"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) start(transporter commonTransporter.Transporter) error {
	log.Println(1)
	array := strings.Split(transporter.Payload.(commonTransporter.RecorderStart).PathOut, "/")
	valid := false
	var path string
	log.Println(2)

	if len(array) == 0 || array == nil {
		return fmt.Errorf("failed to get path out array")
	}

	log.Println(3)
	if len(array) == 1 {
		valid = true
		log.Println(4)
	} else {
		log.Println(5)
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
		valid = true
	}
	log.Println(6)
	if !valid {
		return fmt.Errorf("invalid path: %+v", transporter)
	}

	log.Println(7)
	tempBaseFileName := strings.Split(array[len(array)-1], ".")
	if len(tempBaseFileName) < 2 || tempBaseFileName == nil {
		return fmt.Errorf("failed to get base file name: %s", tempBaseFileName)
	}
	log.Println(8)
	baseFileNameArray := tempBaseFileName[0 : len(tempBaseFileName)-1]
	var baseFileName string
	for _, s := range baseFileNameArray {
		baseFileName += s
	}
	log.Println(9)

	streamIn := fmt.Sprintf("%s://%s%s", v.Config.StreamServerScheme, v.Config.StreamServer, transporter.Payload.(commonTransporter.RecorderStart).StreamIn)
	path = v.Config.RecordingLocation + path

	log.Println(10)
	finish := make(chan bool)

	err := v.cache.Add(fmt.Sprintf("%s_%s", transporter.Unique, finishChannelNameAppend), finish, cache.NoExpiration)
	if err != nil {
		return fmt.Errorf("failed to add finishing channel to cache: %w", err)
	}

	log.Println(11)
	go func() {
		log.Println(14)
		var i uint64
		for {
			v.cache.Delete(transporter.Unique)
			select {
			case <-finish:
				return
			default:
				log.Println(15)
				// Checking if file exists
				_, err = os.Stat(fmt.Sprintf("'%s%s_%d.mkv'", path, baseFileName, i))
				if err == nil {
					break
				}
				log.Println(16)
				err = v.helperStart(transporter, streamIn, path, baseFileName, i)
				log.Println(17)
				if err != nil {
					log.Printf("failed to record: %+v", err)
					return
				}
				log.Println(18)
				time.Sleep(500 * time.Millisecond)
			}
			i++
		}
	}()
	log.Println(12)

	go func() {
		log.Println(13)
		for {
			select {
			case <-finish:
				cmd, ok := v.cache.Get(transporter.Unique)
				if !ok {
					log.Println("unable to get cmd from cache")
					return
				}
				c1 := cmd.(*exec.Cmd)
				err = c1.Process.Kill()
				if err != nil {
					log.Printf("failed to kill recorder: %+v", err)
				}
				v.cache.Delete(transporter.Unique)
				return
			default:
				time.Sleep(1 * time.Second) // This is so it doesn't spam constantly and take the entire CPU up
			}
		}
	}()

	log.Printf("started recording: %s", transporter.Unique)

	return nil
}

func (v *Views) helperStart(transporter commonTransporter.Transporter, streamIn, path, baseFileName string, i uint64) error {
	//err := ffmpeg.Input(streamIn).Output(fmt.Sprintf("'%s%s_%d.mkv'", path, baseFileName, i)).Run()
	//if err != nil {
	//	return fmt.Errorf("failed to run ffmpeg: %w", err)
	//}
	_ = ffmpeg.Stream{}
	c := exec.Command("ffmpeg", "-i", streamIn, "-c", "copy", fmt.Sprintf("'%s%s_%d.mkv'", path, baseFileName, i))
	log.Println(c.String())
	err := v.cache.Add(transporter.Unique, c, cache.NoExpiration)
	if err != nil {
		return fmt.Errorf("failed to add command to cache: %w", err)
	}
	var f *os.File
	f, err = os.OpenFile(fmt.Sprintf("/logs/%s.txt", transporter.Unique), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(fmt.Errorf("failed to open file: %w", err))
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	c.Stdout = f
	c.Stderr = f

	if err = c.Run(); err != nil {
		log.Println("could not run command: ", err)
	}
	return nil
}
