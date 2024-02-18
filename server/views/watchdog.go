package views

import (
	"log"
	"time"
)

func (v *Views) InitWatchdog() *Watchdog {
	return &Watchdog{
		conf:  v.conf,
		store: v.store,
	}
}

func (w *Watchdog) Begin() {
	go func() {
		for {
			if w.conf.Verbose {
				log.Printf("watchdog called")
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
