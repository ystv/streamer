package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/patrickmn/go-cache"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) start(transporter commonTransporter.Transporter) error {
	streamIn := fmt.Sprintf("rtmp://%s%s", v.Config.StreamServer, transporter.Payload.(commonTransporter.ForwarderStart).StreamIn)

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
					c := exec.Command("ffmpeg", "-i", fmt.Sprintf("\"%s\"", streamIn), "-c", "copy", "-f", "flv", fmt.Sprintf("\"%slive/%s\"", v.Config.StreamServer, transporter.Payload.(commonTransporter.ForwarderStart).WebsiteOut), ">>", fmt.Sprintf("\"/logs/%s_0.txt\"", transporter.Unique), "2>&1")
					err = v.cache.Add(transporter.Unique+"0", c, cache.NoExpiration)
					if err != nil {
						log.Printf("failed to add command to cache: %+v", err)
						return
					}
					if err = c.Run(); err != nil {
						log.Println("could not run command: ", err)
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
					c := exec.Command("ffmpeg", "-i", fmt.Sprintf("\"%s\"", streamIn), "-c", "copy", "-f", "flv", fmt.Sprintf("\"%s\"", transporter.Payload.(commonTransporter.ForwarderStart).Streams[j]), ">>", fmt.Sprintf("\"/logs/%s_%d.txt\"", transporter.Unique, j+1), "2>&1")
					err = v.cache.Add(fmt.Sprintf("%s_%d", transporter.Unique, j+1), c, cache.NoExpiration)
					if err != nil {
						log.Println(err)
						return
					}
					if err = c.Run(); err != nil {
						log.Println("could not run command: ", err)
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
