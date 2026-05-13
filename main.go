package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/jcancelli/hot-reload/server"
	"github.com/jcancelli/hot-reload/util"
	watch "github.com/jcancelli/hot-reload/watcher"
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
	config.FillInDefaults()

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)

	watcher, err := watch.NewWatcher(config.Watcher)
	if err != nil {
		log.Fatalln(
			util.WrapError("failed to initialize watcher", err),
		)
	}

	server, err := server.NewServer(config.Server, args.LogHTTP)
	if err != nil {
		log.Fatalln(
			util.WrapError("failed to initialize server", err),
		)
	}

	watcher.Start()
	server.Start()

	<-interrupt
	log.Println("shutting down")
	server.Stop()
	watcher.Stop()
}
