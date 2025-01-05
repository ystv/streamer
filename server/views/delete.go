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

		var response struct {
			Deleted bool   `json:"deleted"`
			Error   string `json:"error"`
		}

		unique := c.FormValue("unique")
		if len(unique) != 10 {
			log.Printf("unique key invalid: %s", unique)
			response.Error = "unique key invalid: " + unique
			return c.JSON(http.StatusOK, response)
		}

		stored, err := v.store.FindStored(unique)
		if err != nil {
			log.Printf("unable to find unique code for delete: %s, %+v", unique, err)
			response.Error = fmt.Sprintf("unable to find unique code for delete: %s, %+v", unique, err)
			return c.JSON(http.StatusOK, response)
		}

		if stored == nil {
			log.Printf("failed to get stored as data is empty")
			response.Error = "failed to get stored as data is empty"
			return c.JSON(http.StatusOK, response)
		}

		err = v.store.DeleteStored(unique)
		if err != nil {
			log.Printf("failed to delete stored: %+v, unique: %s", err, unique)
			response.Error = fmt.Sprintf("failed to delete stored: %+v, unique: %s", err, unique)
			return c.JSON(http.StatusOK, response)
		}

		log.Printf("deleted stored: %s", unique)

		response.Deleted = true
		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
