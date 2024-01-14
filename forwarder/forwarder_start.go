package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) start(transporter commonTransporter.Transporter) error {
	streamIn := fmt.Sprintf("%s%s", v.Config.StreamServer, transporter.Payload.(commonTransporter.ForwarderStart).StreamIn)

	if len(transporter.Payload.(commonTransporter.ForwarderStart).WebsiteOut) > 0 {
		finish := make(chan bool)

		err := v.cache.Add(fmt.Sprintf("%s_0_%s", transporter.Unique, finishChannelNameAppend), finish, cache.NoExpiration)
		if err != nil {
			return fmt.Errorf("failed to add finishing channel to cache: %w", err)
		}

		go func() {
			for {
				v.cache.Delete(fmt.Sprintf("%s_0", transporter.Unique))
				select {
				case <-finish:
					return
				default:
					err = v.helperStart(transporter, streamIn, fmt.Sprintf("%slive/%s", v.Config.StreamServer, transporter.Payload.(commonTransporter.ForwarderStart).WebsiteOut), 0)
					if err != nil {
						log.Printf("failed to stream: %+v", err)
						return
					}
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()

		go func() {
			for {
				select {
				case <-finish:
					cmd, ok := v.cache.Get(fmt.Sprintf("%s_0", transporter.Unique))
					if !ok {
						log.Println("unable to get cmd from cache")
					}
					c1 := cmd.(*exec.Cmd)
					err = c1.Process.Kill()
					if err != nil {
						log.Printf("failed to kill forwarder: %+v", err)
					}
					v.cache.Delete(fmt.Sprintf("%s_0", transporter.Unique))
					return
				default:
					time.Sleep(1 * time.Second) // This is so it doesn't spam constantly and take the entire CPU up
				}
			}
		}()
	}

	for i := 0; i < len(transporter.Payload.(commonTransporter.ForwarderStart).Streams); i++ {
		finish := make(chan bool)

		err := v.cache.Add(fmt.Sprintf("%s_%d_%s", transporter.Unique, i+1, finishChannelNameAppend), finish, cache.NoExpiration)
		if err != nil {
			return fmt.Errorf("failed to add finishing channel to cache: %w", err)
		}

		k := i
		go func() {
			j := k
			for {
				v.cache.Delete(fmt.Sprintf("%s_%d", transporter.Unique, j+1))
				select {
				case <-finish:
					return
				default:
					err = v.helperStart(transporter, streamIn, fmt.Sprintf("\"%s\"", streamIn), j+1)
					if err != nil {
						log.Printf("failed to stream: %+v", err)
						return
					}
					time.Sleep(500 * time.Millisecond)
				}
			}
		}()

		go func() {
			for {
				select {
				case <-finish:
					cmd, ok := v.cache.Get(fmt.Sprintf("%s_%d", transporter.Unique, k))
					if !ok {
						log.Println("unable to get cmd from cache")
						break
					}
					c1 := cmd.(*exec.Cmd)
					err = c1.Process.Kill()
					if err != nil {
						log.Println(err)
					}
					v.cache.Delete(fmt.Sprintf("%s_%d", transporter.Unique, k))
					return
				default:
					time.Sleep(1 * time.Second) // This is so it doesn't spam constantly and take the entire CPU up
				}
			}
		}()
	}

	log.Printf("started forwarder: %s", transporter.Unique)

	return nil
}

func (v *Views) helperStart(transporter commonTransporter.Transporter, streamIn, streamOut string, i int) error {
	c := exec.Command("ffmpeg", "-i", streamIn, "-c", "copy", "-f", "flv", streamOut)
	err := v.cache.Add(transporter.Unique+strconv.Itoa(i), c, cache.NoExpiration)
	if err != nil {
		return fmt.Errorf("failed to add command to cache: %w", err)
	}
	var f *os.File
	f, err = os.OpenFile(fmt.Sprintf("/logs/%s_%d.txt", transporter.Unique, i), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
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
