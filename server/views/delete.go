package views

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

// DeleteFunc will delete the saved stream before it can start
func (v *Views) DeleteFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		if v.conf.Verbose {
			fmt.Println("Delete POST called")
		}

		stored, err := v.store.FindStored(c.FormValue("unique"))
		if err != nil {
			return err
		}

		if stored == nil {
			fmt.Println("no data in stored")
			fmt.Println("REJECTED!")
			return c.String(http.StatusOK, "REJECTED!")
		}

		fmt.Println("DELETED!")
		return c.String(http.StatusOK, "DELETED!")
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
