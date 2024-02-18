package main

import (
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/ystv/streamer/server/store"
	"github.com/ystv/streamer/server/views"
)

type (
	Router struct {
		config views.Config
		router *echo.Echo
		views  *views.Views
	}
)

var (
	verbose bool
	Version = "unknown"
)

//go:embed public/*
var embeddedFiles embed.FS

// main function is the start and the root for the website
func main() {
	if strings.Contains(os.Args[0], "/var/folders") || strings.Contains(os.Args[0], "/tmp/go") || strings.Contains(os.Args[0], "./streamer") {
		if len(os.Args) > 2 {
			log.Fatalf("Arguments error")
		}
		for i := 0; i < len(os.Args)-1; i++ {
			os.Args[i] = os.Args[i+1]
		}
	} else {
		if len(os.Args) > 1 {
			log.Fatalf("Arguments error")
		}
	}
	if os.Args[0] == "-v" {
		verbose = true
	} else {
		verbose = false
	}

	err := godotenv.Load()
	if err != nil {
		log.Printf("error loading .env file: %s", err)
	}

	var config views.Config
	err = envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("failed to process env vars: %s", err)
	}

	root := false

	_, err = os.ReadFile("/bin/streamer")
	if err == nil {
		root = true
	}

	newStore, err := store.NewStore(root)
	if err != nil {
		log.Fatal("Failed to create store: ", err)
	}

	config.Version = Version

	r := &Router{
		config: config,
		router: echo.New(),
		views:  views.New(config, newStore),
	}
	r.router.HideBanner = true

	r.router.Debug = verbose

	r.middleware()

	r.loadRoutes()

	log.Printf("streamer server version: %s", Version)

	r.router.Logger.Error(r.router.Start(r.config.ServerAddress))
	log.Fatalf("failed to start router on address %s", r.config.ServerAddress)
}

func (r *Router) middleware() {
	r.router.Pre(middleware.RemoveTrailingSlash())
	r.router.Use(middleware.Recover())
	r.router.Use(middleware.BodyLimit("15M"))
	r.router.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
}

func (r *Router) loadRoutes() {
	r.router.RouteNotFound("/*", r.views.Error404)

	r.router.HTTPErrorHandler = r.views.CustomHTTPErrorHandler

	assetHandler := http.FileServer(http.FS(echo.MustSubFS(embeddedFiles, "public/")))

	r.router.GET("/public/*", echo.WrapHandler(http.StripPrefix("/public/", assetHandler)))

	validMethods := []string{http.MethodGet, http.MethodPost}
	r.router.Match(validMethods, "/", r.views.HomeFunc)
	r.router.Match(validMethods, "/endpoints", r.views.EndpointsFunc)
	r.router.Match(validMethods, "/streams", r.views.StreamsFunc)                       // Call made by home to view all active streams for the endpoints
	r.router.Match(validMethods, "/start", r.views.StartFunc)                           // Call made by home to start forwarding
	r.router.Match(validMethods, "/resume", r.views.ResumeFunc)                         // To return to the page that controls a stream
	r.router.Match(validMethods, "/status", r.views.StatusFunc)                         // Call made by home to view status
	r.router.Match(validMethods, "/stop", r.views.StopFunc)                             // Call made by home to stop forwarding
	r.router.Match(validMethods, "/list", r.views.ListFunc)                             // List view of current forwards
	r.router.Match(validMethods, "/save", r.views.SaveFunc)                             // Where you can save a stream for later
	r.router.Match(validMethods, "/recall", r.views.RecallFunc)                         // Where you can recall a saved stream to modify it if needed and start it
	r.router.Match(validMethods, "/delete", r.views.DeleteFunc)                         // Deletes the saved stream if it is no longer needed
	r.router.Match(validMethods, "/startUnique", r.views.StartUniqueFunc)               // Call made by home to start forwarding from a recalled stream
	r.router.Match(validMethods, "/youtubehelp", r.views.YoutubeHelpFunc)               // YouTube help page
	r.router.Match(validMethods, "/facebookhelp", r.views.FacebookHelpFunc)             // Facebook help page
	r.router.Match(validMethods, "/"+r.config.StreamerWebsocketPath, r.views.Websocket) // Websocket for the recorder and forwarder to communicate on
	r.router.Match(validMethods, "/activeStreams", r.views.ActiveStreamsFunc)
	r.router.GET("/api/health", func(c echo.Context) error {
		marshal, err := json.Marshal(struct {
			Status int `json:"status"`
		}{
			Status: http.StatusOK,
		})
		if err != nil {
			log.Printf("failed to marshal api health: %+v", err)
			return &echo.HTTPError{
				Code:     http.StatusBadRequest,
				Message:  err.Error(),
				Internal: err,
			}
		}

		c.Response().Header().Set("Content-Type", "application/json")
		return c.JSON(http.StatusOK, marshal)
	})
}
