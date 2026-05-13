package watcher

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"text/template"
)

var (
	cmdArgsTemplate = template.New("Command Arguments")
)

type FileHandler interface {
	FileHandlerTrigger

	Execute(filePath string) error
}

type fileHandler struct {
	FileHandler

	config   FileHandlerConfig
	triggers []FileHandlerTrigger
	cmd      []string
}

func NewFileHandler(config FileHandlerConfig) (FileHandler, error) {
	self := &fileHandler{}
	self.config = config
	self.cmd = config.Command
	self.triggers = make([]FileHandlerTrigger, 0)
	if config.Files != nil {
		trigger, err := newFileListHandlerTrigger(config)
		if err != nil {
			return nil, err
		}
		self.triggers = append(self.triggers, trigger)
	}
	if config.Extensions != nil {
		trigger, err := newExtensionFileHandlerTrigger(config)
		if err != nil {
			return nil, err
		}
		self.triggers = append(self.triggers, trigger)
	}
	if config.Regex != "" {
		trigger, err := newRegexFileHandlerTrigger(config)
		if err != nil {
			return nil, err
		}
		self.triggers = append(self.triggers, trigger)
	}
	return self, nil
}

func (self *fileHandler) IsTriggeredBy(filePath string) bool {
	for _, trigger := range self.triggers {
		if trigger.IsTriggeredBy(filePath) {
			return true
		}
	}
	return false
}

func (self *fileHandler) Execute(filePath string) error {
	log.Printf("[WATCHER] [%s] -------------", self.config.Name)

	params, err := NewCommandParams(filePath)
	if err != nil {
		return err
	}
	cmd, err := compileCommand(self.cmd, params)
	if err != nil {
		return err
	}
	command := exec.Command(cmd[0], cmd[1:]...)
	if self.config.PrintStdOut {
		command.Stdout = os.Stdout
	}
	if self.config.PrintStdErr {
		command.Stderr = os.Stderr
	}
	if err := command.Run(); err != nil {
		return err
	}

	return nil
}

// Values that will be replaced into a command
type CommandParams struct {
	FileBase string
	FileAbs  string
	FileDir  string
}

func NewCommandParams(fileName string) (CommandParams, error) {
	var err error
	self := CommandParams{}
	self.FileBase = filepath.Base(fileName)
	self.FileAbs, err = filepath.Abs(fileName)
	self.FileDir = filepath.Dir(self.FileAbs)
	return self, err
}

// Replace template variables with their value inside a command
func compileCommand(cmd []string, params CommandParams) ([]string, error) {
	cmdCopy := slices.Clone(cmd)
	var buff bytes.Buffer
	for i, arg := range cmdCopy {
		t, err := cmdArgsTemplate.Parse(arg)
		if err != nil {
			return nil, err
		}
		err = t.Execute(&buff, params)
		if err != nil {
			return nil, err
		}
		cmdCopy[i] = buff.String()
		buff.Reset()
	}
	return cmdCopy, nil
}
