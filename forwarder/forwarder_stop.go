package main

import (
	"fmt"
	"strings"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) stop(transporter commonTransporter.Transporter) error {
	found := false
	for k, item := range v.cache.Items() {
		if strings.Contains(k, transporter.Unique) && strings.Contains(k, finishChannelNameAppend) {
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
