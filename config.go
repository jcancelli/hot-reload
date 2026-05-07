package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
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
	ConfigPath string
}

func ParseCliArgs() (args CliArgs, err error) {
	flag.StringVar(
		&args.ConfigPath,
		"c",
		"",
		"Path to the configuration file",
	)

	flag.Parse()

	if args.ConfigPath == "" {
		return args, errors.New("missing config file path")
	}

	if err = CheckValidFile(args.ConfigPath); err != nil {
		return args, errors.Join(
			err,
			errors.New("invalid config file path"),
		)
	}

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
