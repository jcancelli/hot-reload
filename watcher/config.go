package watcher

const (
	// The default value of the time in milliseconds that elapses between each time the fs events are polled (and the handlers triggered)
	DefaultPollIntervalMs = 1000
)

type Config struct {
	// The directories that will be watched for file events
	Directories []string `json:"directories"`
	// The time in milliseconds that elapses between each time the fs events are polled (and the handlers triggered)
	PollIntervalMs uint `json:"pollIntervalMs"`
	// The handlers that will listen for file events
	Handlers []FileHandlerConfig `json:"handlers"`
}

// Set the values that were omitted to their default values
func (self *Config) FillInDefaults() {
	if self.PollIntervalMs == 0 {
		self.PollIntervalMs = DefaultPollIntervalMs
	}
	for i := range self.Handlers {
		self.Handlers[i].FillInDefaults()
	}
}

// Configuration for a file handler
type FileHandlerConfig struct {
	// The name of this handler
	Name string `json:"name"`
	// The command that will be executed when this handler is triggered
	Command []string `json:"command"`
	// A list of extensions that will trigger this handler
	Extensions []string `json:"extensions"`
	// A list of files that will be watched for changes
	Files []string `json:"files"`
	// A regex that will trigger this handler
	Regex string `json:"regex"`
	// Wether stdout of the command should be written to stdout
	PrintStdOut bool `json:"printStdOut"`
	// Wether stderr of the command should be written to stderr
	PrintStdErr bool `json:"printStdErr"`
}

func (self *FileHandlerConfig) FillInDefaults() {
	if self.Name == "" {
		self.Name = "unnamed-handler"
	}
}
