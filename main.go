package main

import (
	_ "embed"
	"fmt"
	"log"

	"github.com/jcancelli/hot-reload/server"
	"github.com/jcancelli/hot-reload/util"
)

var (
	//go:embed skeleton-config.json
	skeletonConfig string
)

func main() {
	args, err := ParseCliArgs()
	if err != nil {
		log.Fatalln(
			util.WrapError("failed to parse cli arguments", err),
		)
	}

	if args.DumpSkeletonConfig {
		fmt.Print(skeletonConfig)
		return
	}

	config, err := LoadConfig(args.ConfigPath)
	if err != nil {
		log.Fatalln(
			util.WrapError("failed to load configuration", err),
		)
	}

	watcher, err := NewWatcher(config.Watch)
	if err != nil {
		log.Fatalln(
			util.WrapError("failed to initialize watcher", err),
		)
	}

	server, err := server.NewServer(config.Server)
	if err != nil {
		log.Fatalln(
			util.WrapError("failed to initialize server", err),
		)
	}

	go func() {
		if err := watcher.Start(); err != nil {
			log.Fatalln(err)
		}
	}()
	if err = server.Start(); err != nil {
		log.Fatalln(err.Error())
	}
}
