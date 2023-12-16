package views

import (
	"fmt"
)

// SavedStreamCheck checks if there are any existing streams still registered in the database
func (v *Views) SavedStreamCheck() bool {
	if v.conf.Verbose {
		fmt.Println("Saved Stream Check called")
	}

	stored, err := v.store.GetStored()
	if err != nil {
		fmt.Println(err)
		return false
	}

	return len(stored) > 0
}
