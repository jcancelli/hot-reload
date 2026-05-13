package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jcancelli/hot-reload/util"
	"github.com/radovskyb/watcher"
)

var (
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
	}
)

type WebsocketServer struct {
	http.Handler

	fileEvent    <-chan watcher.Event
	shutdown     <-chan bool
	pingInterval time.Duration
	pongWait     time.Duration
}

func NewWebsocketServer(
	fileEvent <-chan watcher.Event,
	shutdown <-chan bool,
	pingIntervalMs uint,
) (self *WebsocketServer, err error) {
	self = &WebsocketServer{
		fileEvent:    fileEvent,
		shutdown:     shutdown,
		pingInterval: time.Millisecond * time.Duration(pingIntervalMs),
		pongWait:     time.Millisecond * time.Duration(pingIntervalMs*5),
	}
	return
}

func (self *WebsocketServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	ws, err := wsUpgrader.Upgrade(response, request, nil)
	if err != nil {
		log.Println(
			util.WrapError("[WS] failed to upgrade connection", err),
		)
		return
	}
	defer func() {
		ws.Close()
		log.Println("[WS] client disconnected")
	}()

	log.Println("[WS] client connected")

	ws.SetReadDeadline(time.Now().Add(self.pongWait))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(self.pongWait))
		return nil
	})

	// Reader
	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				if err != websocket.ErrCloseSent && websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
				) {
					log.Printf("[WS] unexpected error: %v", err)
				}
				return
			}
		}
	}()

	ping := time.NewTicker(self.pingInterval)
	defer ping.Stop()

	// Writer
	for {
		select {
		case <-ping.C:
			err := ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
			if err != nil {
				if err != websocket.ErrCloseSent && websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
				) {
					log.Printf("[WS] unexpected error: %v", err)
				}
				break
			}

		case <-self.shutdown:
			if err := ws.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(
					websocket.CloseNormalClosure,
					"",
				),
			); err != nil {
				if websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
				) {
					log.Printf("[WS] unexpected error: %v", err)
				}
			}
			break

		case <-self.fileEvent:
			log.Println("[WS] reload signaled")
			if err := ws.WriteMessage(websocket.TextMessage, []byte("RELOAD")); err != nil {
				if err != websocket.ErrCloseSent && websocket.IsUnexpectedCloseError(
					err,
					websocket.CloseNormalClosure,
					websocket.CloseGoingAway,
				) {
					log.Printf("[WS] unexpected error: %v", err)
				}
				break
			}
		}
	}
}
