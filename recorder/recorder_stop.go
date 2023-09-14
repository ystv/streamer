package main

import (
	"fmt"
)

func (v *Views) stop(transporter Transporter) error {
	finish, ok := v.cache.Get(transporter.Unique + "Finish")
	if !ok {
		return fmt.Errorf("unable to find channel: %s", transporter.Unique)
	}
	close(finish.(chan struct{}))
	return nil
}
