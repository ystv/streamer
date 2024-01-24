package views

import (
	"log"
)

// ExistingStreamCheck checks if there are any existing streams still registered in the database
func (v *Views) ExistingStreamCheck() bool {
	if v.conf.Verbose {
		log.Println("Existing Stream Check called")
	}

	streams, err := v.store.GetStreams()
	if err != nil {
		log.Printf("failed to get streams for existingStreamCheck: %+v", err)
		return false
	}

	stored, err := v.store.GetStored()
	if err != nil {
		log.Printf("failed to get stored for existingStreamCheck: %+v", err)
		return false
	}

	return len(streams) > 0 || len(stored) > 0
}
