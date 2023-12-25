package views

import (
	"encoding/json"
	"fmt"
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

	clientChannel := make(chan []byte)
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

	loop := true

	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		_ = ws.Close()
	}()

	for loop {
		select {
		case res := <-clientChannel:
			err = ws.WriteMessage(websocket.TextMessage, res)
			if err != nil {
				log.Printf("failed to write response: %+v, server %s", err, responseTransporter.Server)
				close(internalChannel)
				close(clientChannel)
				v.cache.Delete(responseTransporter.Server.String())
				v.cache.Delete(responseTransporter.Server.String() + internalChannelNameAppend)
				loop = false
			}

			_, msg, err = ws.ReadMessage()
			if err != nil {
				log.Printf("failed to read message: %+v, server: %s", err, responseTransporter.Server)
				close(internalChannel)
				close(clientChannel)
				v.cache.Delete(responseTransporter.Server.String())
				v.cache.Delete(responseTransporter.Server.String() + internalChannelNameAppend)
				loop = false
			}
			internalChannel <- msg
			log.Printf("Message received from \"%s\": %s", responseTransporter.Server, msg)
		case <-ticker.C:
			err = ws.WriteMessage(websocket.TextMessage, []byte(specialWSMessage.Ping))
			if err != nil {
				log.Printf("failed to write ping for %s: %+v", responseTransporter.Server, err)
				close(internalChannel)
				close(clientChannel)
				v.cache.Delete(responseTransporter.Server.String())
				v.cache.Delete(responseTransporter.Server.String() + internalChannelNameAppend)
				loop = false
			}
			var msgType int
			msgType, msg, err = ws.ReadMessage()
			if err != nil || msgType != websocket.TextMessage || string(msg) != specialWSMessage.Pong.String() {
				log.Printf("failed to read pong for %s: %+v", responseTransporter.Server, err)
				close(internalChannel)
				close(clientChannel)
				v.cache.Delete(responseTransporter.Server.String())
				v.cache.Delete(responseTransporter.Server.String() + internalChannelNameAppend)
				loop = false
			}
		}
	}
	return nil
}
