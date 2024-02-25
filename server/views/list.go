package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/server/templates"
)

type (
	listedStream struct {
		Code  string `json:"code"`
		Input string `json:"input"`
	}
)

// ListFunc lists all current streams that are registered in the database
func (v *Views) ListFunc(c echo.Context) error {
	/*if !authenticate(w, r) {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("error loading .env file: %s", err)
		}

		jwtAuthentication := os.Getenv("JWT_AUTHENTICATION")

		http.Redirect(w, r, jwtAuthentication+"list", http.StatusTemporaryRedirect)
		return
	}*/
	if c.Request().Method == "GET" {
		if v.conf.Verbose {
			log.Println("Stop GET called")
		}

		return v.template.RenderTemplate(c.Response().Writer, nil, templates.ListTemplate)
	} else if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("Stop POST called")
		}

		var response struct {
			ActiveList []listedStream `json:"activeList"`
			SavedList  []listedStream `json:"savedList"`
			Error      string         `json:"error"`
		}

		response.ActiveList = []listedStream{}
		response.SavedList = []listedStream{}

		streams, err := v.store.GetStreams()
		if err != nil {
			log.Printf("failed to get streams: %+v", err)
			response.Error = fmt.Sprintf("failed to get streams: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		for _, s := range streams {
			response.ActiveList = append(response.ActiveList, listedStream{
				Code:  s.Stream,
				Input: s.Input,
			})
		}

		stored, err := v.store.GetStored()
		if err != nil {
			log.Printf("failed to get stored: %+v", err)
			response.Error = fmt.Sprintf("failed to get stored: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		for _, s := range stored {
			response.SavedList = append(response.SavedList, listedStream{
				Code:  s.Stream,
				Input: s.Input,
			})
		}

		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
