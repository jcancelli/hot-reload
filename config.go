package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/jcancelli/hot-reload/server"
)

type Config struct {
	Server server.Config `json:"server"`
	Watch  WatchConfig   `json:"watcher"`
}

type WatchConfig struct {
	Directories []string            `json:"directories"`
	Handlers    []FileHandlerConfig `json:"handlers"`
}

type FileHandlerConfig struct {
	Pattern string   `json:"pattern"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Events  []string `json:"events"`
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

	flag.Parse()

	return args, nil
}
