package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"time"

	"github.com/radovskyb/watcher"
)

type Watcher struct {
	*watcher.Watcher
	Handlers []FileHandler
}

func NewWatcher(config WatchConfig) (self Watcher, err error) {
	for _, fileWatcherConfig := range config.Handlers {
		fileWatcher, err := NewFileHandler(fileWatcherConfig)
		if err != nil {
			return self, err
		}
		self.Handlers = append(self.Handlers, fileWatcher)
	}

	self.Watcher = watcher.New()

	for _, directory := range config.Directories {
		err = self.AddRecursive(directory)
		if err != nil {
			return
		}
	}

	return self, err
}

func (self *Watcher) Start() error {
	go func() {
		for {
			select {
			case event := <-self.Event:
				for _, handler := range self.Handlers {
					if handler.IsHandlerFor(event) {
						if err := handler.Run(event); err != nil {
							log.Println(err)
						}
					}
				}
			case err := <-self.Error:
				log.Fatalln(err)
			case <-self.Closed:
				return
			}
		}
	}()
	return self.Watcher.Start(time.Millisecond * 200)
}

type FileHandler struct {
	Pattern *regexp.Regexp
	Command string
	Args    []string
	Events  map[watcher.Op]bool
}

func NewFileHandler(config FileHandlerConfig) (self FileHandler, err error) {
	self.Pattern, err = regexp.Compile(config.Pattern)
	if err != nil {
		return
	}

	self.Command = config.Command
	self.Args = config.Args

	self.Events = make(map[watcher.Op]bool)
	for _, event := range config.Events {
		switch event {
		case watcher.Create.String():
			self.Events[watcher.Create] = true
		case watcher.Write.String():
			self.Events[watcher.Write] = true
		case watcher.Remove.String():
			self.Events[watcher.Remove] = true
		case watcher.Rename.String():
			self.Events[watcher.Rename] = true
		case watcher.Chmod.String():
			self.Events[watcher.Chmod] = true
		case watcher.Move.String():
			self.Events[watcher.Move] = true
		default:
			err = fmt.Errorf("invalid file handler event: %s", event)
			return
		}
	}

	return
}

func (self *FileHandler) IsHandlerFor(event watcher.Event) bool {
	if !self.Events[event.Op] {
		return false
	}
	return self.Pattern.Match([]byte(event.OldPath)) || self.Pattern.Match([]byte(event.Path))
}

func (self *FileHandler) Run(event watcher.Event) error {
	return exec.Command(self.Command, self.Args...).Run()
}
