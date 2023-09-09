package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/mapstructure"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

type (
	Transporter struct {
		Action  string      `json:"action"`
		Unique  string      `json:"unique"`
		Payload interface{} `json:"payload"`
	}

	ForwarderStart struct {
		StreamIn   string            `json:"streamIn"`
		WebsiteOut string            `json:"websiteOut"`
		Streams    map[string]string `json:"streams"`
	}

	ForwarderStatus struct {
		Website bool `json:"website"`
		Streams int  `json:"streams"`
	}

	RecorderStart struct {
		StreamIn string `json:"streamIn"`
		PathOut  string `json:"pathOut"`
	}

	Config struct {
		StreamServer          string `envconfig:"STREAM_SERVER"`
		RecordingLocation     string `envconfig:"RECORDING_LOCATION"`
		StreamerWebAddress    string `envconfig:"STREAMER_WEB_ADDRESS"`
		StreamerWebsocketPath string `envconfig:"STREAMER_WEBSOCKET_PATH"`
	}
)

func main() {
	//fmt.Println("echo", os.Args)
	//if strings.Contains(os.Args[0], "/var/folders") || strings.Contains(os.Args[0], "/tmp/go") || strings.Contains(os.Args[0], "./recorder_start") {
	//	if len(os.Args) != 7 && len(os.Args) != 3 {
	//		fmt.Println("echo " + string(rune(len(os.Args))))
	//		log.Fatalf("echo Arguments error")
	//	}
	//	for i := 0; i < len(os.Args)-1; i++ {
	//		os.Args[i] = os.Args[i+1]
	//	}
	//} else {
	//	if len(os.Args) != 6 && len(os.Args) != 2 {
	//		fmt.Println("echo " + string(rune(len(os.Args))))
	//		log.Fatalf("echo Arguments error")
	//	}
	//}
	//method := os.Args[0]
	//switch method {
	//case "start":
	//	start(os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5])
	//	break
	//case "stop":
	//	stop(os.Args[1])
	//	break
	//case "status":
	//	status(os.Args[1])
	//	break
	//default:
	//	log.Fatalf("echo Invalid method used: %s", method)
	//}
	_ = godotenv.Load(".env")

	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatalf("failed to process env vars: %s", err)
	}

	for {
		run(config)
	}
}

func run(config Config) {
	messageOut := make(chan []byte)
	errorChannel := make(chan error, 1)
	done := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		for sig := range interrupt {
			fmt.Printf("signal: %s\n", sig)
			os.Exit(0)
		}
	}()
	u := url.URL{Scheme: "ws", Host: config.StreamerWebAddress, Path: "/" + config.StreamerWebsocketPath}
	log.Printf("connecting to %s://%s", u.Scheme, u.Host)
	c, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		if resp != nil {
			log.Printf("handshake failed with status %d", resp.StatusCode)
		}
		log.Printf("failed to dial url: %+v", err)
	}

	//finish := false
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Restarting...")
			//finish = true
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

	//When the program closes close the connection
	defer c.Close()
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				close(errorChannel)
			}
		}()
		err = c.WriteMessage(websocket.TextMessage, []byte("recorder"))
		if err != nil {
			log.Printf("failed to write name: %+v", err)
			//finish = true
			close(errorChannel)
			return
		}

		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("failed to read acknowledgement: %+v", err)
			close(errorChannel)
			//finish = true
			return
		}

		if string(msg) != "ACKNOWLEDGED" {
			log.Printf("failed to read acknowledgement: %s", string(msg))
			close(errorChannel)
			//finish = true
			return
		} else {
			log.Println("ACKNOWLEDGED")
			log.Printf("connected to  %s://%s", u.Scheme, u.Host)
		}
		//for !finish {
		for {
			//if finish {
			//	close(errorChannel)
			//	return
			//}
			msgType, message, err := c.ReadMessage()
			//fmt.Printf("Message type: %d\nMessage: %s\nError: %+v\n", msgType, message, err)
			if err != nil {
				log.Printf("failed to read: %+v", err)
				close(errorChannel)
				return
			}
			if msgType == websocket.TextMessage && string(message) == "ping" {
				err = c.WriteMessage(websocket.TextMessage, []byte("pong"))
				if err != nil {
					log.Printf("failed to write pong: %+v", err)
					close(errorChannel)
					return
				}
				continue
			}
			log.Printf("Received message: %s", message)
			messageOut <- message
		}
	}()

	//ticker := time.NewTicker(time.Second)
	//defer ticker.Stop()
	defer close(errorChannel)
	//for !finish {
	for {
		//if finish {
		//	//ticker.Stop()
		//	return
		//}
		select {
		case <-done:
		case <-interrupt:
		case <-errorChannel:
			return
		case m := <-messageOut:
			log.Printf("Picked up message %s", m)

			var t Transporter

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

			switch t.Action {
			case "start":
				fmt.Println(t)

				var t1 RecorderStart

				err = mapstructure.Decode(t.Payload, &t1)
				if err != nil {
					log.Printf("failed to decode: %+v", err)
				}

				t.Payload = t1

				fmt.Println(t)
			case "status":
			case "stop":
			default:
				log.Printf("failed to get action: %s", t.Action)
				err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("failed to get action: %s", t.Action)))
				if err != nil {
					log.Printf("failed to write error response : %+v", err)
					return
				}
				continue
			}
			//err := c.WriteMessage(websocket.TextMessage, []byte(m))
			//if err != nil {
			//	log.Println("write2:", err)
			//	return
			//}
			//case t := <-ticker.C:
			//	log.Printf("Ticker: %s", t.Format("2006-01-02T15:04:05"))
			//	err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			//	if err != nil {
			//		log.Println("write3:", err)
			//		return
			//	}
			//case <-interrupt:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			//err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			//if err != nil {
			//	log.Println("write close:", err)
			//	return
			//}
			//select {
			//case <-done:
			//}
			//return
			//case <-errorChannel:
			//	//finish = true
			//	return
		}
	}
}
