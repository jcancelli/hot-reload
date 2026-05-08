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

	fileChangeEvent <-chan watcher.Event
	pingInterval    time.Duration
}

func NewWebsocketServer(fileChangeEvent <-chan watcher.Event, pingIntervalMs uint) (self *WebsocketServer, err error) {
	self = &WebsocketServer{
		fileChangeEvent: fileChangeEvent,
		pingInterval:    time.Millisecond * time.Duration(pingIntervalMs),
	}
	return
}

func (self *WebsocketServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	ws, err := wsUpgrader.Upgrade(response, request, nil)
	if err != nil {
		log.Fatalln(
			util.WrapError("failed to upgrade websocket connection", err),
		)
	}
	log.Println("hot-reload client connected")
	go self.writer(ws)
	self.reader(ws)
}

func (self *WebsocketServer) writer(ws *websocket.Conn) {
	defer ws.Close()

	pingTicker := time.NewTicker(self.pingInterval)
	defer pingTicker.Stop()

	for {
		var err error

		select {
		case <-self.fileChangeEvent:
			err = ws.WriteJSON(NewReloadMessage())

		case <-pingTicker.C:
			err = ws.WriteJSON(NewPingMessage())
		}

		if err != nil {
			if websocket.IsCloseError(err) {
				log.Println("hot-reload client disconnected")
				break
			}
			log.Println(
				util.WrapError("error while sending message to client", err),
			)
		}
	}
}

func (self *WebsocketServer) reader(ws *websocket.Conn) {
	// Nothing for now
}

type MessageKind string

const (
	PingMsgKind   MessageKind = "ping"
	ReloadMsgKind MessageKind = "reload"
)

type Message struct {
	Kind MessageKind `json:"kind"`
}

type PingMessage struct {
	Message
}

func NewPingMessage() PingMessage {
	return PingMessage{
		Message: Message{
			Kind: PingMsgKind,
		},
	}
}

type ReloadMessage struct {
	Message
}

func NewReloadMessage() ReloadMessage {
	return ReloadMessage{
		Message: Message{
			Kind: ReloadMsgKind,
		},
	}
}
