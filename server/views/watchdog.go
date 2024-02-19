package views

import (
	"log"
	"time"
)

func (v *Views) BeginWatchdog() {
	log.Printf("watchdog initiailised")
	go func() {
		for {
			if w.conf.Verbose {
				log.Printf("watchdog called")
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
