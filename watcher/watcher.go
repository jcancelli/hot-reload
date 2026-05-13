package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	w "github.com/radovskyb/watcher"
)

type Watcher interface {
	Start() error
	Stop() error
}

type watcher struct {
	Watcher

	shutdown             chan bool
	w                    *w.Watcher
	pollIntervalDuration time.Duration
	pollInterval         *time.Ticker
	mux                  sync.Mutex
	eventQueue           []w.Event
	handlers             []FileHandler
}

func NewWatcher(config Config) (Watcher, error) {
	self := &watcher{}

	self.shutdown = make(chan bool)

	dirs := make([]string, 0)
	for _, dir := range config.Directories {
		// Ensure absolute path
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}
		// Ensure is directory
		stat, err := os.Stat(absDir)
		if err != nil {
			return nil, err
		}
		if !stat.IsDir() {
			return nil, fmt.Errorf("path %s is not a directory", absDir)
		}
		dirs = append(dirs, dir)
	}

	self.w = w.New()
	for _, dir := range dirs {
		if err := self.w.AddRecursive(dir); err != nil {
			return nil, err
		}
	}

	self.pollIntervalDuration = time.Duration(config.PollIntervalMs) * time.Millisecond
	self.pollInterval = time.NewTicker(self.pollIntervalDuration)

	self.eventQueue = make([]w.Event, 0)

	self.handlers = make([]FileHandler, len(config.Handlers))
	for i, handlerConfig := range config.Handlers {
		handler, err := NewFileHandler(handlerConfig)
		if err != nil {
			return nil, err
		}
		self.handlers[i] = handler
	}

	return self, nil
}

func (self *watcher) Start() error {
	go func() {
	loop:
		for {
			select {
			case <-self.shutdown:
				break loop

			case event := <-self.w.Event:
				self.mux.Lock()
				self.eventQueue = append(self.eventQueue, event)
				self.mux.Unlock()

			case <-self.w.Closed:
				break loop

			case err := <-self.w.Error:
				log.Printf("[WATCHER] %v", err)
				break loop
			}
		}
	}()
	go func() {
		if err := self.w.Start(self.pollIntervalDuration); err != nil {
			log.Printf("[WATCHER] %v", err)
		}
	}()
	go func() {
	loop:
		for {
			select {
			case <-self.shutdown:
				break loop

			default:
				self.mux.Lock()
				for _, event := range self.eventQueue {
					for _, handler := range self.handlers {
						if handler.IsTriggeredBy(event.Path) {
							if err := handler.Execute(event.Path); err != nil {
								log.Printf("[WATCHER] %v", err)
							}
						}
					}
				}
				clear(self.eventQueue)
				self.mux.Unlock()
			}
		}
	}()

	return nil
}

func (self *watcher) Stop() error {
	self.w.Close()
	close(self.shutdown)
	return nil
}
