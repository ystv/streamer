package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/common/transporter/server"
)

func (v *Views) ServerErrorFunc(c echo.Context) error {
	if c.Request().Method == "POST" {
		unique := c.FormValue("unique")

		var response struct {
			ServerError string `json:"serverError"`
			Error       string `json:"error"`
		}

		recorder := true

		if len(unique) != 0 && len(unique) != 10 {
			log.Printf("unique key invalid: %s", unique)
			response.Error = fmt.Sprintf("unique key invalid: %s<br>", unique)
			return c.JSON(http.StatusOK, response)
		} else if len(unique) == 10 {
			stream, err := v.store.FindStream(unique)
			if err != nil {
				log.Printf("unable to find unique code for serverError: %s, %+v", unique, err)
				response.Error = fmt.Sprintf("unable to find unique code for serverError: %s, %+v<br>", unique, err)
				return c.JSON(http.StatusOK, response)
			}

			if stream == nil {
				log.Printf("failed to get stream as data is empty")
				response.Error = "failed to get stream as data is empty<br>"
				return c.JSON(http.StatusOK, response)
			}

			recorder = len(stream.Recording) > 0
		}

		_, rec := v.cache.Get(server.Recorder.String())
		_, fow := v.cache.Get(server.Forwarder.String())

		var errString string

		errExtra := ", this may be temporary, if this persists for more than a minute then please contact <code>#computing</code> on Slack<br>"

		if !rec && !fow && recorder {
			errString = "No recorder or forwarder available" + errExtra
		} else if !rec && recorder {
			errString = "No recorder available" + errExtra
		} else if !fow {
			errString = "No forwarder available" + errExtra
		}

		response.ServerError = errString
		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
