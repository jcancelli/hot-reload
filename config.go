package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
)

type Config struct {
	Serve ServeConfig `json:"serve"`
	Watch WatchConfig `json:"watch"`
}

type ServeConfig struct {
	Port              uint   `json:"port"`
	Directory         string `json:"directory"`
	ClientScriptRoute string `json:"clientScriptRoute"`
	WebsocketRoute    string `json:"websocketRoute"`
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

	if err = CheckValidDirectory(config.Serve.Directory); err != nil {
		return config, errors.Join(err, errors.New("invalid serve directory"))
	}
	for _, directory := range config.Watch.Directories {
		if err = CheckValidDirectory(directory); err != nil {
			return config, errors.Join(err, errors.New("invalid watch directory"))
		}
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

func CheckValidDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%v is not a directory", path)
	}
	return nil
}

func CheckValidFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%v is a directory", path)
	}
	return nil
}
