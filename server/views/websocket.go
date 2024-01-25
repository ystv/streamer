package views

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	commonTransporter "github.com/ystv/streamer/common/transporter"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"

	"github.com/ystv/streamer/common/transporter/server"
	specialTransporter "github.com/ystv/streamer/common/transporter/special"
	specialWSMessage "github.com/ystv/streamer/common/wsMessages/special"
)

var wsUpgrade = websocket.Upgrader{}

func (v *Views) Websocket(c echo.Context) error {
	ws, err := wsUpgrade.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return fmt.Errorf("failed to upgrade web socket: %w", err)
	}
	defer func(ws *websocket.Conn) {
		_ = ws.Close()
	}(ws)

	_, msg, err := ws.ReadMessage()
	if err != nil {
		log.Printf("failed to read message from websocket: %+v", err)
		_ = ws.Close()
		return nil
	}

	var responseTransporter specialTransporter.InitiationTransporter

	err = json.Unmarshal(msg, &responseTransporter)
	if err != nil {
		log.Printf("failed to unmarshal response: %+v", err)
		_ = ws.Close()
		return nil
	}

	if responseTransporter.Server != server.Forwarder && responseTransporter.Server != server.Recorder {
		log.Printf("failed connecting %s, invalid name", responseTransporter.Server)
		_ = ws.Close()
		return nil
	}

	log.Println("connecting", responseTransporter.Server)

	if responseTransporter.Version != v.conf.Version {
		log.Printf("%s has a version mismatch, server version: %s, %s version: %s", responseTransporter.Server, v.conf.Version, responseTransporter.Server, responseTransporter.Version)
	}

	clientChannel := make(chan TransporterRouter)
	internalChannel := make(chan []byte)

	err = v.cache.Add(responseTransporter.Server.String(), clientChannel, cache.NoExpiration)
	if err != nil {
		log.Printf("failed to add channel to cache: %+v, server: %s", err, responseTransporter.Server)
		_ = ws.Close()
		return nil
	}

	err = v.cache.Add(responseTransporter.Server.String()+internalChannelNameAppend, internalChannel, cache.NoExpiration)
	if err != nil {
		log.Printf("failed to add finish channel to cache: %+v, server: %s", err, responseTransporter.Server)
		_ = ws.Close()
		return nil
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte(specialWSMessage.Acknowledged))
	if err != nil {
		log.Printf("failed to write acknowledgement response: %+v, server: %s", err, responseTransporter.Server)
		_ = ws.Close()
		return nil
	}

	log.Println("connected", responseTransporter.Server)

	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		_ = ws.Close()
	}()

	for {
		select {
		case res := <-clientChannel:
			var transportUniqueReturning commonTransporter.TransporterUnique
			err = v.cache.Add(res.TransporterUnique.ID, res.ReturningChannel, cache.DefaultExpiration)
			if err != nil {
				log.Printf("failed to add id to cache: %+v", err)
			}

			res.ReturningChannel = nil

			var send []byte
			send, err = json.Marshal(res.TransporterUnique)
			if err != nil {
				log.Printf("failed to marshal transportUnique: %+v", err)
			}

			err = ws.WriteMessage(websocket.TextMessage, send)
			if err != nil {
				log.Printf("failed to write message: %+v, server %s", err, responseTransporter.Server)
				close(internalChannel)
				close(clientChannel)
				v.cache.Delete(responseTransporter.Server.String())
				v.cache.Delete(responseTransporter.Server.String() + internalChannelNameAppend)
				return nil
			}

			_, msg, err = ws.ReadMessage()
			if err != nil {
				log.Printf("failed to read message: %+v, server %s", err, responseTransporter.Server)
				close(internalChannel)
				close(clientChannel)
				v.cache.Delete(responseTransporter.Server.String())
				v.cache.Delete(responseTransporter.Server.String() + internalChannelNameAppend)
				return nil
			}

			err = json.Unmarshal(msg, &transportUniqueReturning)
			if err != nil {
				log.Printf("failed to unmarshal message: %+v, server %s", err, responseTransporter.Server)
				close(internalChannel)
				close(clientChannel)
				v.cache.Delete(responseTransporter.Server.String())
				v.cache.Delete(responseTransporter.Server.String() + internalChannelNameAppend)
				return nil
			}

			returnChannel, ok := v.cache.Get(transportUniqueReturning.ID)
			if !ok {
				log.Printf("failed to find channel for server %s", responseTransporter.Server)
				continue
			}

			var receive []byte
			switch transportUniqueReturning.Payload.(type) {
			case string:
				receive = []byte(transportUniqueReturning.Payload.(string))
				break
			case commonTransporter.ResponseTransporter:
				receive, err = json.Marshal(transportUniqueReturning.Payload)
				if err != nil {
					log.Printf("failed to unmarshal response: %+v, server %s", err, responseTransporter.Server)
					close(internalChannel)
					close(clientChannel)
					v.cache.Delete(responseTransporter.Server.String())
					v.cache.Delete(responseTransporter.Server.String() + internalChannelNameAppend)
					return nil
				}
				log.Printf("Message received from %s: %s", responseTransporter.Server, msg)
				break
			default:
				log.Printf("invalid returning message: %#v, server %s", transportUniqueReturning.Payload, responseTransporter.Server)
				close(internalChannel)
				close(clientChannel)
				v.cache.Delete(responseTransporter.Server.String())
				v.cache.Delete(responseTransporter.Server.String() + internalChannelNameAppend)
				return nil
			}

			returnChannel.(chan []byte) <- receive
			break
		case <-ticker.C:
			go func() {
				returningChannel := make(chan []byte)

				sendingTransporter := commonTransporter.TransporterUnique{
					ID:               uuid.NewString(),
					Payload:          specialWSMessage.Ping,
					ReturningChannel: returningChannel,
				}

				clientChannel <- sendingTransporter

				received := <-returningChannel

				if string(received) != specialWSMessage.Pong.String() {
					log.Printf("failed to read pong for %s: %s", responseTransporter.Server, string(received))
					close(internalChannel)
					close(clientChannel)
					v.cache.Delete(responseTransporter.Server.String())
					v.cache.Delete(responseTransporter.Server.String() + internalChannelNameAppend)
				}
			}()
		}
	}
}
