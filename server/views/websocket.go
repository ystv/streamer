package views

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"log"
	"time"
)

var wsUpgrade = websocket.Upgrader{}

func (v *Views) Websocket(c echo.Context) error {
	ws, err := wsUpgrade.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer func(ws *websocket.Conn) {
		_ = ws.Close()
	}(ws)

	_, msg, err := ws.ReadMessage()
	if err != nil {
		c.Logger().Error(err)
		_ = ws.Close()
		return nil
	}

	name := string(msg)

	log.Println("connecting", name)

	clientChannel := make(chan []byte)
	internalChannel := make(chan []byte)

	err = v.cache.Add(name, clientChannel, cache.NoExpiration)
	if err != nil {
		c.Logger().Error(err)
		_ = ws.Close()
		return nil
	}

	err = v.cache.Add(name+internalChannelNameAppend, internalChannel, cache.NoExpiration)
	if err != nil {
		c.Logger().Error(err)
		_ = ws.Close()
		return nil
	}

	err = ws.WriteMessage(websocket.TextMessage, []byte("ACKNOWLEDGED"))
	if err != nil {
		c.Logger().Error(err)
		_ = ws.Close()
		return nil
	}

	log.Println("connected", name)

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
				c.Logger().Error(err)
				close(internalChannel)
				close(clientChannel)
				v.cache.Delete(name)
				v.cache.Delete(name + internalChannelNameAppend)
				loop = false
			}

			_, msg, err = ws.ReadMessage()
			if err != nil {
				c.Logger().Error(err)
				internalChannel <- []byte(fmt.Sprintf("ERROR: failed to read message from %s: %+v", name, err))
				close(internalChannel)
				close(clientChannel)
				v.cache.Delete(name)
				v.cache.Delete(name + internalChannelNameAppend)
				loop = false
			}
			internalChannel <- msg
			log.Printf("Message received from \"%s\": %s", name, msg)
		case <-ticker.C:
			err = ws.WriteMessage(websocket.TextMessage, []byte("ping"))
			if err != nil {
				log.Printf("failed to write ping for %s: %+v", name, err)
				close(clientChannel)
				v.cache.Delete(name)
				v.cache.Delete(name + internalChannelNameAppend)
				loop = false
			}
			msgType, msg, err := ws.ReadMessage()
			if err != nil || msgType != websocket.TextMessage || string(msg) != "pong" {
				log.Printf("failed to read pong for %s: %+v", name, err)
				close(clientChannel)
				v.cache.Delete(name)
				v.cache.Delete(name + internalChannelNameAppend)
				loop = false
			}
		}
	}
	return nil
}
