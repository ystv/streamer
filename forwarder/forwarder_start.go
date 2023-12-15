package main

import (
	"github.com/patrickmn/go-cache"
	"log"
	"os/exec"
	"strconv"
	"time"
)

func (v *Views) start(transporter Transporter) error {
	streamIn := "rtmp://" + v.Config.StreamServer + transporter.Payload.(ForwarderStart).StreamIn

	if len(transporter.Payload.(ForwarderStart).WebsiteOut) > 0 {
		finish := make(chan bool)

		err := v.cache.Add(transporter.Unique+"_0Finish", finish, cache.NoExpiration)
		if err != nil {
			return err
		}

		go func() {
			for {
				v.cache.Delete(transporter.Unique + "_0")
				switch {
				case <-finish:
					return
				default:
					c := exec.Command("ffmpeg", "-i", "\""+streamIn+"\"", "-c", "copy", "-f", "flv", "\""+v.Config.StreamServer+"live/"+transporter.Payload.(ForwarderStart).WebsiteOut+"\"", ">>", "\"/logs/"+transporter.Unique+"_0.txt\"", "2>&1")
					err = v.cache.Add(transporter.Unique+"0", c, cache.NoExpiration)
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
				switch {
				case <-finish:
					cmd, ok := v.cache.Get(transporter.Unique + "_0")
					if !ok {
						log.Println("unable to get cmd from cache")
					}
					c1 := cmd.(*exec.Cmd)
					err = c1.Process.Kill()
					if err != nil {
						log.Println(err)
					}
					v.cache.Delete(transporter.Unique + "_0")
					return
				default:
					time.Sleep(1 * time.Second) // This is so it doesn't spam constantly and take the entire CPU up
				}
			}
		}()
	}

	for i := 0; i < len(transporter.Payload.(ForwarderStart).Streams); i++ {
		finish := make(chan bool)

		err := v.cache.Add(transporter.Unique+"_"+strconv.Itoa(i+1)+"Finish", finish, cache.NoExpiration)
		if err != nil {
			return err
		}

		k := i
		go func() {
			j := k
			for {
				v.cache.Delete(transporter.Unique + "_" + strconv.Itoa(j+1))
				switch {
				case <-finish:
					return
				default:
					c := exec.Command("ffmpeg", "-i", "\""+streamIn+"\"", "-c", "copy", "-f", "flv", "\""+transporter.Payload.(ForwarderStart).Streams[j]+"\"", ">>", "\"/logs/"+transporter.Unique+"_"+strconv.Itoa(j+1)+".txt\"", "2>&1")
					err = v.cache.Add(transporter.Unique+"_"+strconv.Itoa(j+1), c, cache.NoExpiration)
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
				switch {
				case <-finish:
					cmd, ok := v.cache.Get(transporter.Unique + "_" + strconv.Itoa(k))
					if !ok {
						log.Println("unable to get cmd from cache")
					}
					c1 := cmd.(*exec.Cmd)
					err = c1.Process.Kill()
					if err != nil {
						log.Println(err)
					}
					v.cache.Delete(transporter.Unique + "_" + strconv.Itoa(k))
					return
				default:
					time.Sleep(1 * time.Second) // This is so it doesn't spam constantly and take the entire CPU up
				}
			}
		}()
	}

	return nil
}
