package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// StartFunc is the core of the program, where it takes the values set by the user in the webpage and processes the data and sends it to the recorder and the forwarder
func (v *Views) StartFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Start POST called")
		}

		var response struct {
			Unique string `json:"unique"`
			Error  string `json:"error"`
		}

		unique, err := v.generateUnique()
		if err != nil {
			log.Printf("failed to get unique: %+v", err)
			response.Error = fmt.Sprintf("failed to get unique: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		err = v.startingWSHelper(c, unique, nonStoredStart)
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
