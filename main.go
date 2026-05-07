package main

import (
	_ "embed"
	"fmt"
	"log"
)

var (
	//go:embed skeleton-config.json
	skeletonConfig string
)

func main() {
	args, err := ParseCliArgs()
	if err != nil {
		log.Fatalf("failed to parse cli arguments\n", err.Error())
	}

	if args.DumpSkeletonConfig {
		fmt.Print(skeletonConfig)
		return
	}

	config, err := LoadConfig(args.ConfigPath)
	if err != nil {
		log.Fatalf("failed to load config file: %s\n", err.Error())
	}

	watcher, err := NewWatcher(config.Watch)
	if err != nil {
		log.Fatalf("failed to init watcher: %s\n", err.Error())
	}

	server, err := NewServer(config.Serve)
	if err != nil {
		log.Fatalf("failed to init server: %s\n", err.Error())
	}

	log.Printf("listening on port %d\n", config.Serve.Port)
	go func() {
		if err := watcher.Start(); err != nil {
			log.Fatalln(err)
		}
	}()
	if err = server.Start(); err != nil {
		log.Fatalln(err.Error())
	}
}
