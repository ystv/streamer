package views

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ystv/streamer/server/templates"
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
			log.Println("List GET called")
		}

		data := struct {
			ActivePage string
		}{
			ActivePage: "list",
		}

		return v.template.RenderTemplate(c.Response().Writer, data, templates.ListTemplate)
	} else if c.Request().Method == "POST" {
		if v.conf.Verbose {
			log.Println("List POST called")
		}

		var response struct {
			ActiveList []ListedStream `json:"activeList"`
			SavedList  []ListedStream `json:"savedList"`
			Error      string         `json:"error"`
		}

		response.ActiveList = []ListedStream{}
		response.SavedList = []ListedStream{}

		streams, err := v.store.GetStreams()
		if err != nil {
			log.Printf("failed to get streams: %+v", err)
			response.Error = fmt.Sprintf("failed to get streams: %+v", err)
			return c.JSON(http.StatusOK, response)
		}

		for _, s := range streams {
			response.ActiveList = append(response.ActiveList, ListedStream{
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
			response.SavedList = append(response.SavedList, ListedStream{
				Code:  s.Stream,
				Input: s.Input,
			})
		}

		return c.JSON(http.StatusOK, response)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, "invalid method")
}
