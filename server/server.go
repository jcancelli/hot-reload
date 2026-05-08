package server

import (
	"context"
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
	fileEvent    chan watcher.Event
	shutdown     chan bool
}

// Create a new server
func NewServer(config Config, logHttp bool) (self *Server, err error) {
	self = &Server{}

	self.fileEvent = make(chan watcher.Event)
	self.shutdown = make(chan bool)

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
	rootHandler := middleware.NoCache(
		http.FileServer(
			http.Dir(config.Directory),
		),
	)
	if logHttp {
		rootHandler = middleware.LogRequest(rootHandler)
	}
	multiplexer.Handle("/", rootHandler)

	multiplexer.HandleFunc(config.ClientScriptRoute, func(response http.ResponseWriter, request *http.Request) {
		response.Header().Add("Content-Type", "text/javascript")
		if _, err := response.Write(self.clientScript); err != nil {
			log.Printf("failed to write client script response: %s", err.Error())
		}
	})

	wsServer, err := NewWebsocketServer(self.fileEvent, self.shutdown, config.PingIntervalMs)
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
func (self *Server) Start() {
	// Watcher
	go func() {
		for {
			select {
			case ev := <-self.watcher.Event:
				self.fileEvent <- ev

			case err := <-self.watcher.Error:
				log.Println(err)
				self.watcher.Close()
				return

			case <-self.watcher.Closed:
				return
			}
		}
	}()
	go func() {
		if err := self.watcher.Start(time.Second); err != nil {
			log.Println(err)
		}
	}()

	// HTTP server
	go func() {
		log.Printf("listening on %s\n", self.server.Addr)
		if err := self.server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Println(err)
			}
		}
	}()
}

func (self *Server) Stop() {
	self.watcher.Close()
	close(self.shutdown)
	self.server.Shutdown(context.Background())
}
