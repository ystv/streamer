package views

import (
	"fmt"
)

// ActiveStreamCheck checks if there are any existing streams still registered in the database
func (v *Views) ActiveStreamCheck() bool {
	if v.conf.Verbose {
		fmt.Println("Active Stream Check called")
	}

	streams, err := v.store.GetStreams()
	if err != nil {
		fmt.Println(err)
		return false
	}

	return len(streams) > 0
}
