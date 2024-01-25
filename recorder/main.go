package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/patrickmn/go-cache"

	commonTransporter "github.com/ystv/streamer/common/transporter"
	"github.com/ystv/streamer/common/transporter/server"
	specialTransporter "github.com/ystv/streamer/common/transporter/special"
	"github.com/ystv/streamer/common/wsMessages"
	specialWSMessage "github.com/ystv/streamer/common/wsMessages/special"
)

type (
	Config struct {
		StreamServer            string `envconfig:"STREAM_SERVER"`
		StreamServerScheme      string `envconfig:"STREAM_SERVER_SCHEME"`
		RecordingLocation       string `envconfig:"RECORDING_LOCATION"`
		StreamerWebAddress      string `envconfig:"STREAMER_WEB_ADDRESS"`
		StreamerWebsocketPath   string `envconfig:"STREAMER_WEBSOCKET_PATH"`
		StreamerWebsocketScheme string `envconfig:"STREAMER_WEBSOCKET_SCHEME"`
	}

	Views struct {
		Config Config
		cache  *cache.Cache
	}
)

const finishChannelNameAppend = "Finish"

var Version = "unknown"

func main() {
	_ = godotenv.Load(".env")

	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("failed to process env vars: %s", err)
	}

	err = os.MkdirAll("/logs", 0777)
	if err != nil {
		log.Fatalf("error creating /logs: %+v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.GET("/api/health", func(c echo.Context) error {
		var marshal []byte
		marshal, err = json.Marshal(struct {
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

	go func() {
		if err = e.Start(":1323"); err != nil {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		for sig := range interrupt {
			if err = e.Shutdown(context.Background()); err != nil {
				e.Logger.Fatal(err)
			}
			log.Printf("signal: %s\n", sig)
			os.Exit(0)
		}
	}()

	v := Views{
		Config: config,
		cache:  cache.New(cache.NoExpiration, 1*time.Hour),
	}

	for {
		log.Printf("streamer recorder version: %s\n", Version)
		v.run(config, interrupt)
		time.Sleep(5 * time.Second)
	}
}

func (v *Views) run(config Config, interrupt chan os.Signal) {
	messageOut := make(chan commonTransporter.TransporterUnique)
	errorChannel := make(chan error, 1)
	done := make(chan struct{})
	u := url.URL{Scheme: config.StreamerWebsocketScheme, Host: config.StreamerWebAddress, Path: "/" + config.StreamerWebsocketPath}
	log.Printf("connecting to %s://%s", u.Scheme, u.Host)
	c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		if resp != nil {
			log.Printf("handshake failed with status %d", resp.StatusCode)
		}
		log.Printf("failed to dial url: %+v", err)
	}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Restarting...")
			select {
			case <-messageOut:
				break
			default:
				close(messageOut)
			}
			select {
			case <-errorChannel:
				break
			default:
				close(errorChannel)
			}
			select {
			case <-done:
				break
			default:
				close(done)
			}
			time.Sleep(5 * time.Second)
			return
		}
	}()

	defer func(c *websocket.Conn) {
		_ = c.Close()
	}(c)
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				close(errorChannel)
			}
		}()
		defer func(c *websocket.Conn) {
			_ = c.Close()
		}(c)
		response := specialTransporter.InitiationTransporter{
			Server:  server.Recorder,
			Version: Version,
		}

		var resBytes []byte
		resBytes, err = json.Marshal(response)
		if err != nil {
			log.Printf("failed to marshal initial: %+v", err)
			close(errorChannel)
			return
		}

		err = c.WriteMessage(websocket.TextMessage, resBytes)
		if err != nil {
			log.Printf("failed to write name and version: %+v", err)
			close(errorChannel)
			return
		}

		var msg []byte
		_, msg, err = c.ReadMessage()
		if err != nil {
			log.Printf("failed to read acknowledgement: %+v", err)
			close(errorChannel)
			return
		}

		if string(msg) != specialWSMessage.Acknowledged.String() {
			log.Printf("failed to read acknowledgement: %s", string(msg))
			close(errorChannel)
			return
		}

		log.Printf("connected to %s://%s", u.Scheme, u.Host)

		for {
			var msgType int
			var message []byte
			msgType, message, err = c.ReadMessage()
			if err != nil {
				log.Printf("failed to read: %+v", err)
				log.Printf("message type: %d", msgType)
				log.Printf("message contents: %s", string(message))
				close(errorChannel)
				return
			}

			var receivedMessage commonTransporter.TransporterUnique
			err = json.Unmarshal(message, &receivedMessage)
			if err != nil {

			}

			switch receivedMessage.Payload.(type) {
			case commonTransporter.Transporter:
				break
			case string:
				if msgType == websocket.TextMessage && receivedMessage.Payload.(string) == specialWSMessage.Ping.String() {
					receivedMessage.Payload = specialWSMessage.Pong

					var responsePing []byte
					responsePing, err = json.Marshal(receivedMessage)
					err = c.WriteMessage(websocket.TextMessage, responsePing)
					if err != nil {
						log.Printf("failed to write pong: %+v", err)
						close(errorChannel)
						return
					}
					continue
				}
			}
			//if msgType == websocket.TextMessage && string(message) == specialWSMessage.Ping.String() {
			//	err = c.WriteMessage(websocket.TextMessage, []byte(specialWSMessage.Pong))
			//	if err != nil {
			//		log.Printf("failed to write pong: %+v", err)
			//		close(errorChannel)
			//		return
			//	}
			//	continue
			//}
			log.Printf("Received message: %s", message)
			messageOut <- receivedMessage
		}
	}()

	defer close(errorChannel)
	for {
		select {
		case <-done:
		case <-interrupt:
		case <-errorChannel:
			return
		case m := <-messageOut:
			log.Printf("Picked up message %s", m)

			t := m.Payload.(commonTransporter.Transporter)

			if len(t.Unique) != 10 {
				kill := v.errorResponse(fmt.Errorf("failed to get unique, length is not equal to 10: %d", len(t.Unique)), c)
				if kill {
					return
				}
				continue
			}

			var out string
			switch t.Action {
			case "start":
				var t1 commonTransporter.RecorderStart

				err = mapstructure.Decode(t.Payload, &t1)
				if err != nil {
					kill := v.errorResponse(fmt.Errorf("failed to decode: %w", err), c)
					if kill {
						return
					}
					continue
				}

				if len(t1.StreamIn) == 0 || len(t1.PathOut) == 0 {
					kill := v.errorResponse(fmt.Errorf("failed to get payload for start: %+v", t1), c)
					if kill {
						return
					}
					continue
				}

				t.Payload = t1

				err = v.start(t)
				if err != nil {
					kill := v.errorResponse(fmt.Errorf("failed to start recorder: %w", err), c)
					if kill {
						return
					}
					continue
				}
			case "status":
				out, err = v.status(t)
				if err != nil {
					kill := v.errorResponse(fmt.Errorf("failed to get status recorder: %w", err), c)
					if kill {
						return
					}
					continue
				}
			case "stop":
				err = v.stop(t)
				if err != nil {
					kill := v.errorResponse(fmt.Errorf("failed to stop recorder: %w", err), c)
					if kill {
						return
					}
					continue
				}
			default:
				kill := v.errorResponse(fmt.Errorf("failed to get action: %s", t.Action), c)
				if kill {
					return
				}
				continue
			}

			response := commonTransporter.ResponseTransporter{Status: wsMessages.Okay}

			if len(out) > 0 {
				response.Payload = out
			}

			m.Payload = response

			var resBytes []byte
			resBytes, err = json.Marshal(m)
			if err != nil {
				log.Printf("failed to marshal response: %+v", err)
			}

			err = c.WriteMessage(websocket.TextMessage, resBytes)
			if err != nil {
				log.Printf("failed to write okay response : %+v", err)
				return
			}
		}
	}
}

func (v *Views) errorResponse(incomingErr error, c *websocket.Conn) bool {
	log.Print(incomingErr)
	response := commonTransporter.ResponseTransporter{
		Status:  wsMessages.Error,
		Payload: incomingErr.Error(),
	}

	var resBytes []byte
	resBytes, err := json.Marshal(response)
	if err != nil {
		log.Printf("failed to marshal response: %+v", err)
	}

	err = c.WriteMessage(websocket.TextMessage, resBytes)
	if err != nil {
		log.Printf("failed to write error response : %+v", err)
		return true
	}
	return false
}
