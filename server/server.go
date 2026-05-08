package server

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jcancelli/hot-reload/server/middleware"
	"github.com/jcancelli/hot-reload/util"
	"github.com/radovskyb/watcher"
)

// A server is responsible for watching and serving the files inside a directry and notifying the
// client for changes to such files
type Server struct {
	watcher      *watcher.Watcher
	server       http.Server
	clientScript []byte
	event        chan watcher.Event
}

// Create a new server
func NewServer(config Config) (self *Server, err error) {
	self = &Server{}

	self.event = make(chan watcher.Event)

	config.FillInDefaults()
	if err = config.Validate(); err != nil {
		return nil, util.WrapError("invalid server config", err)
	}

	self.watcher = watcher.New()
	if err = self.watcher.AddRecursive(config.Directory); err != nil {
		return
	}
	self.watcher.SetMaxEvents(1)

	if self.clientScript, err = config.GenerateClientScript(); err != nil {
		return nil, util.WrapError("failed to generate client script", err)
	}

	multiplexer := http.NewServeMux()
	multiplexer.Handle("/", middleware.LogRequest(
		middleware.NoCache(
			http.FileServer(
				http.Dir(config.Directory),
			),
		),
	))
	multiplexer.HandleFunc(config.ClientScriptRoute, func(response http.ResponseWriter, request *http.Request) {
		response.Header().Add("Content-Type", "text/javascript")
		if _, err := response.Write(self.clientScript); err != nil {
			log.Printf("failed to write client script response: %s", err.Error())
		}
	})

	wsServer, err := NewWebsocketServer(self.event, config.PingIntervalMs)
	if err != nil {
		return nil, util.WrapError("failed to create websocket server", err)
	}
	multiplexer.Handle(config.WebSocketRoute, wsServer)

	self.server = http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: multiplexer,
	}

	return
}

// Start the server
func (self *Server) Start() error {
	go func(event chan watcher.Event) {
		for {
			select {
			case ev := <-self.watcher.Event:
				event <- ev
			case err := <-self.watcher.Error:
				log.Fatalln(err)
			case <-self.watcher.Closed:
				break
			}
		}
	}(self.event)
	go func() {
		if err := self.watcher.Start(time.Second); err != nil {
			log.Fatalln(err)
		}
	}()
	return self.server.ListenAndServe()
}
