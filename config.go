package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/jcancelli/hot-reload/server"
	"github.com/jcancelli/hot-reload/watcher"
)

type Config struct {
	Server  server.Config  `json:"server"`
	Watcher watcher.Config `json:"watcher"`
}

func (self *Config) FillInDefaults() {
	self.Server.FillInDefaults()
	self.Watcher.FillInDefaults()
}

func LoadConfig(path string) (Config, error) {
	var config Config

	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

type CliArgs struct {
	ConfigPath         string
	DumpSkeletonConfig bool
	LogHTTP            bool
}

func ParseCliArgs() (args CliArgs, err error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return args, fmt.Errorf("failed to get working directory. %w", err)
	}

	flag.StringVar(
		&args.ConfigPath,
		"c",
		path.Join(workingDir, "hot-reload.json"),
		"Path to the configuration file",
	)
	flag.BoolVar(
		&args.DumpSkeletonConfig,
		"skeleton-config",
		false,
		"Print to standard out a skeleton for a hot-reload configuration file",
	)
	flag.BoolVar(
		&args.LogHTTP,
		"log-http",
		false,
		"Log to standard out incoming HTTP requests",
	)

	flag.Parse()

	return args, nil
}
