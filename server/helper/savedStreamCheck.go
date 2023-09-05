package helper

import (
	"database/sql"
	"fmt"
)

// SavedStreamCheck checks if there are any existing streams still registered in the database
func SavedStreamCheck(verbose bool) bool {
	if verbose {
		fmt.Println("Saved Stream Check called")
	}
	db, err := sql.Open("sqlite3", "db/streams.db")
	if err != nil {
		fmt.Println(err)
	} else {
		rows, err := db.Query("SELECT stream FROM stored")
		if err != nil {
			fmt.Println(err)
		}

		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}

		var stream string

		for rows.Next() {
			err = rows.Scan(&stream)
			if err != nil {
				fmt.Println(err)
			}
			err = rows.Close()
			if err != nil {
				fmt.Println(err)
			}
			err = db.Close()
			if err != nil {
				fmt.Println(err)
			}
			return true
		}
		err = rows.Close()
		if err != nil {
			fmt.Println(err)
		}
		err = db.Close()
		if err != nil {
			fmt.Println(err)
		}
		return false
	}
	err = db.Close()
	if err != nil {
		fmt.Println(err)
	}
	return false
}
