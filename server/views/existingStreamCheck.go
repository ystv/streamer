package views

import (
	"fmt"
)

// ExistingStreamCheck checks if there are any existing streams still registered in the database
func (v *Views) ExistingStreamCheck() bool {
	if v.conf.Verbose {
		fmt.Println("Existing Stream Check called")
	}

	streams, err := v.store.GetStreams()
	if err != nil {
		fmt.Println(err)
		return false
	}

	stored, err := v.store.GetStored()
	if err != nil {
		fmt.Println(err)
		return false
	}

	return len(streams) > 0 || len(stored) > 0
}
