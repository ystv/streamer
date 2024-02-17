package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// DeleteFunc will delete the saved stream before it can start
func (v *Views) DeleteFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Delete POST called")
		}

		unique := c.FormValue("unique")

		stored, err := v.store.FindStored(unique)
		if err != nil {
			return fmt.Errorf("failed to get stored: %w, unique: %s", err, unique)
		}

		if stored == nil {
			log.Println("no data in stored")
			log.Printf("rejected delete: %s", unique)
			return c.String(http.StatusOK, "REJECTED!")
		}

		err = v.store.DeleteStream(unique)
		if err != nil {
			log.Printf("failed to delete stored: %+v, unique: %s", err, unique)
			return c.String(http.StatusOK, "REJECTED!")
		}

		log.Printf("deleted stored: %s", unique)
		return c.String(http.StatusOK, "DELETED!")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
