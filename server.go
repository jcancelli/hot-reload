package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/jcancelli/hot-reload/middleware"
	"github.com/radovskyb/watcher"
	"golang.org/x/net/websocket"
)

var (
	//go:embed hot-reload.js
	clientScriptRaw string
)

type Server struct {
	watcher           *watcher.Watcher
	server            http.Server
	clientScriptRoute string
	websocketRoute    string
	clientScript      []byte
	reloadChan        chan bool
}

type clientScriptParams struct {
	WebSocketRoute string
	Port           uint
}

func NewServer(config ServeConfig) (self Server, err error) {
	self.reloadChan = make(chan bool)

	self.watcher = watcher.New()
	self.watcher.AddRecursive(config.Directory)

	clientScriptTemplate, err := template.New("client script").Parse(clientScriptRaw)
	if err != nil {
		return
	}
	if config.WebsocketRoute != "" {
		self.websocketRoute = config.WebsocketRoute
	} else {
		self.websocketRoute = "/hot-reload-websocket"
	}

	var clientScriptBuffer bytes.Buffer
	if err = clientScriptTemplate.Execute(&clientScriptBuffer, clientScriptParams{
		WebSocketRoute: self.websocketRoute,
		Port:           config.Port,
	}); err != nil {
		return
	}
	self.clientScript = clientScriptBuffer.Bytes()

	serveMux := http.NewServeMux()

	serveMux.Handle("/", http.FileServer(http.Dir(config.Directory)))
	serveMux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))

	if config.ClientScriptRoute == "" {
		self.clientScriptRoute = "/hot-reload.js"
	} else {
		self.clientScriptRoute = config.ClientScriptRoute
	}
	serveMux.HandleFunc(self.clientScriptRoute, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/javascript")
		_, err := w.Write(self.clientScript)
		if err != nil {
			log.Printf("error while writing response for client script: %s", err.Error())
		}
	})

	serveMux.Handle(self.websocketRoute, websocket.Handler(func(connection *websocket.Conn) {
		go func(connection *websocket.Conn, reload <-chan bool) {
			for range reload {
				_, err := connection.Write([]byte("reload"))
				if err != nil {
					log.Printf("error while notifying websocket: %s", err.Error())
				}
			}
		}(connection, self.reloadChan)
	}))

	self.server = http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: middleware.NoCache(middleware.LogRequest(serveMux)),
	}

	return
}

func (self *Server) Start() error {
	go func() {
		for {
			select {
			case <-self.watcher.Event:
				self.reloadChan <- true
			case err := <-self.watcher.Error:
				log.Fatalf("error while watching serve directory: %s\n", err.Error())
			case <-self.watcher.Closed:
				return
			}
		}
	}()
	go func() {
		if err := self.watcher.Start(time.Second); err != nil {
			log.Fatalf("error while watching serve directory: %s", err.Error())
		}
	}()
	return self.server.ListenAndServe()
}
