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
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/patrickmn/go-cache"

	commonTransporter "github.com/ystv/streamer/common/transporter"
	"github.com/ystv/streamer/common/transporter/server"
	"github.com/ystv/streamer/common/wsMessages"
)

type (
	Config struct {
		StreamServer            string `envconfig:"STREAM_SERVER"`
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
			fmt.Println(err)
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
			fmt.Printf("signal: %s\n", sig)
			os.Exit(0)
		}
	}()

	v := Views{
		Config: config,
		cache:  cache.New(cache.NoExpiration, 1*time.Hour),
	}

	for {
		v.run(config, interrupt)
	}
}

var pinging atomic.Bool

func (v *Views) run(config Config, interrupt chan os.Signal) {
	messageOut := make(chan []byte)
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

	//When the program closes, close the connection
	defer func(c *websocket.Conn) {
		_ = c.Close()
	}(c)
	go func() {
		pinging.Store(false)
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				close(errorChannel)
			}
		}()
		err = c.WriteMessage(websocket.TextMessage, []byte(server.Forwarder))
		if err != nil {
			log.Printf("failed to write name: %+v", err)
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

		if string(msg) != wsMessages.Acknowledged.String() {
			log.Printf("failed to read acknowledgement: %s", string(msg))
			close(errorChannel)
			return
		}
		log.Println(wsMessages.Acknowledged)
		log.Printf("connected to  %s://%s", u.Scheme, u.Host)

		for {
			var msgType int
			var message []byte
			msgType, message, err = c.ReadMessage()
			if err != nil {
				log.Printf("failed to read: %+v", err)
				close(errorChannel)
				return
			}
			if msgType == websocket.TextMessage && string(message) == wsMessages.Ping.String() {
				pinging.Store(true)
				err = c.WriteMessage(websocket.TextMessage, []byte(wsMessages.Pong))
				if err != nil {
					log.Printf("failed to write pong: %+v", err)
					close(errorChannel)
					return
				}
				pinging.Store(false)
				continue
			}
			log.Printf("Received message: %s", message)
			messageOut <- message
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

			var t commonTransporter.Transporter

			err = json.Unmarshal(m, &t)
			if err != nil {
				log.Printf("failed to unmarshal data: %+v", err)
				err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("failed to unmarshal data: %+v", err)))
				if err != nil {
					log.Printf("failed to write error response : %+v", err)
					return
				}
				continue
			}

			if len(t.Unique) != 10 {
				log.Printf("failed to get unique, length is not equal to 10: %d", len(t.Unique))
				err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("failed to get unique, length is not equal to 10: %d", len(t.Unique))))
				if err != nil {
					log.Printf("failed to write error response : %+v", err)
					return
				}
				continue
			}

			var out commonTransporter.ForwarderStatusResponse
			switch t.Action {
			case "start":
				var t1 commonTransporter.ForwarderStart

				err = mapstructure.Decode(t.Payload, &t1)
				if err != nil {
					log.Printf("failed to decode: %+v", err)
					err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("ERROR: failed to decode: %+v", err)))
					if err != nil {
						log.Printf("failed to write error response : %+v", err)
					}
					return
				}

				if len(t1.StreamIn) == 0 || len(t1.Streams) == 0 {
					err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("ERROR: failed to get payload for start: %+v", err)))
					if err != nil {
						log.Printf("failed to write error response : %+v", err)
					}
					return
				}

				t.Payload = t1

				err = v.start(t)
				if err != nil {
					log.Printf("failed to start recorder: %+v", err)
					err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("ERROR: failed to start recorder: %+v", err)))
					if err != nil {
						log.Printf("failed to write error response : %+v", err)
					}
					return
				}
			case "status":
				var t1 commonTransporter.ForwarderStatus

				err = mapstructure.Decode(t.Payload, &t1)
				if err != nil {
					log.Printf("failed to decode: %+v", err)
					err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("ERROR: failed to decode: %+v", err)))
					if err != nil {
						log.Printf("failed to write error response : %+v", err)
					}
					return
				}

				if t1.Streams == 0 {
					err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("ERROR: failed to get payload for start: %+v", err)))
					if err != nil {
						log.Printf("failed to write error response : %+v", err)
					}
					return
				}

				t.Payload = t1

				out, err = v.status(t)
				if err != nil {
					log.Printf("failed to get status forwarder: %+v", err)
					err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("ERROR: failed to get status forwarder: %+v", err)))
					if err != nil {
						log.Printf("failed to write error response : %+v", err)
					}
					return
				}
			case "stop":
				err = v.stop(t)
				if err != nil {
					log.Printf("failed to stop fprwarder: %+v", err)
					err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("ERROR: failed to stop forwarder: %+v", err)))
					if err != nil {
						log.Printf("failed to write error response : %+v", err)
					}
					return
				}
			default:
				log.Printf("failed to get action: %s", t.Action)
				err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("failed to get action: %s", t.Action)))
				if err != nil {
					log.Printf("failed to write error response : %+v", err)
					return
				}
				continue
			}

			response := "OKAY"

			if len(out.Streams) > 0 {
				var b []byte
				b, err = json.Marshal(out)
				if err != nil {
					log.Printf("failed marshaling out: %+v", err)
					return
				}
				response += "±~±" + string(b) // Some arbitrary connector string that is unlikely to be used ever by anything else
			}

			err = c.WriteMessage(websocket.TextMessage, []byte(response))
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
	for pinging.Load() {
		time.Sleep(10 * time.Millisecond)
	}
	err = c.WriteMessage(websocket.TextMessage, resBytes)
	if err != nil {
		log.Printf("failed to write error response : %+v", err)
		return true
	}
	return false
}
