package main

import (
	"fmt"
	"log"
	"strings"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) stop(transporter commonTransporter.Transporter) error {
	found := false
	log.Println(transporter.Unique)
	for k, item := range v.cache.Items() {
		log.Println(k)
		if strings.Contains(k, transporter.Unique) && strings.Contains(k, finishChannelNameAppend) {
			log.Println("Found")
			found = true
			close(item.Object.(chan bool))
			v.cache.Delete(k)
		}
	}
	if !found {
		return fmt.Errorf("unable to find channels for: %s", transporter.Unique)
	}
	return nil
}
