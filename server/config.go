package server

import (
	"bytes"
	_ "embed"
	"os"
	"text/template"

	"github.com/jcancelli/hot-reload/util"
)

const (
	// The default port of the server
	DefaultPort uint = 8080
	// The default route from where the js script that handles page reloads will be served from
	DefaultClientScriptRoute = "/hot-reload.js"
	// The default route from where the websocket that sends reload signals to the client will be listening from
	DefaultWebsocketRoute = "/hot-reload-websocket"
	// The default interval between ping messages sent by the server to the client
	DefaultPingIntervalMs = 10_000
)

var (
	//go:embed hot-reload.js
	clientScriptTemplate string
)

// Returns the default directory that contains the files that will be watched for changes and served by the server
func DefaultDirectory() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", util.WrapError("failed to get working directory", err)
	}
	return currentDir, nil
}

// Configuration for the server
type Config struct {
	// The default port of the server
	Port uint `json:"port"`
	// The directory that contains the files that will be watched for changes and served by the server
	Directory string `json:"directory"`
	// The route from where the js script that handles page reloads will be served from
	ClientScriptRoute string `json:"clientScriptRoute"`
	// The route from where the websocket that sends reload signals to the client will be listening from
	WebSocketRoute string `json:"websocketRoute"`
	// The interval between ping messages sent by the server to the client
	PingIntervalMs uint `json:"pingIntervalMs"`
}

// Sets the uninitialized fields of this config to their default values
func (self *Config) FillInDefaults() {
	if self.Port == 0 {
		self.Port = DefaultPort
	}
	if self.Directory == "" {
		self.Directory, _ = DefaultDirectory()
	}
	if self.ClientScriptRoute == "" {
		self.ClientScriptRoute = DefaultClientScriptRoute
	}
	if self.WebSocketRoute == "" {
		self.WebSocketRoute = DefaultWebsocketRoute
	}
	if self.PingIntervalMs == 0 {
		self.PingIntervalMs = DefaultPingIntervalMs
	}
}

// Validate this config
func (self Config) Validate() error {
	if err := util.CheckValidDirectory(self.Directory); err != nil {
		return util.WrapError("invalid serve directory", err)
	}
	return nil
}

// Generate a js script that will connect to the server launched with this config
func (self Config) GenerateClientScript() ([]byte, error) {
	templ, err := template.New("client script template").Parse(clientScriptTemplate)
	if err != nil {
		return nil, util.WrapError("failed to create client script template", err)
	}

	var buff bytes.Buffer
	if err = templ.Execute(&buff, self); err != nil {
		return nil, util.WrapError("failed to execute client script template", err)
	}
	return buff.Bytes(), nil
}
