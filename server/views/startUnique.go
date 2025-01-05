package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// StartUniqueFunc is the core of the program, where it takes the values set by the user in the webpage and processes the data and sends it to the recorder and the forwarder with a specified unique key
func (v *Views) StartUniqueFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("StartUnique POST called")
		}

		var response struct {
			Unique string `json:"unique"`
			Error  string `json:"error"`
		}

		unique := c.FormValue("unique_code")
		if len(unique) != 10 {
			log.Printf("unique key invalid: %s", unique)
			response.Error = "unique key invalid: " + unique
			return c.JSON(http.StatusOK, response)
		}

		stored, err := v.store.FindStored(unique)
		if err != nil {
			log.Printf("unable to find unique code for startUnique: %s, %+v", unique, err)
			response.Error = fmt.Sprintf("unable to find unique code for startUnique: %s, %+v", unique, err)
			return c.JSON(http.StatusOK, response)
		}

		if stored == nil {
			log.Printf("failed to get stored as data is empty")
			response.Error = "failed to get stored as data is empty"
			return c.JSON(http.StatusOK, response)
		}

		err = v.startingWSHelper(c, unique, storedStart)
		if err != nil {
			log.Printf("failed to start: %+v", err)
			response.Error = fmt.Sprintf("failed to start: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		log.Printf("started stream: %s", unique)

		response.Unique = unique
		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
