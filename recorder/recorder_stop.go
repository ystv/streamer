package main

import (
	"fmt"

	commonTransporter "github.com/ystv/streamer/common/transporter"
)

func (v *Views) stop(transporter commonTransporter.Transporter) error {
	finish, ok := v.cache.Get(fmt.Sprintf("%s_%s", transporter.Unique, finishChannelNameAppend))
	if !ok {
		return fmt.Errorf("unable to find channel: %s", transporter.Unique)
	}
	close(finish.(chan bool))
	v.cache.Delete(fmt.Sprintf("%s_%s", transporter.Unique, finishChannelNameAppend))
	return nil
}
