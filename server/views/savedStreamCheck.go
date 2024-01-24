package views

import (
	"log"
)

// SavedStreamCheck checks if there are any existing streams still registered in the database
func (v *Views) SavedStreamCheck() bool {
	if v.conf.Verbose {
		log.Println("Saved Stream Check called")
	}

	stored, err := v.store.GetStored()
	if err != nil {
		log.Printf("failed to get stored for saveStreamCheck: %+v", err)
		return false
	}

	return len(stored) > 0
}
