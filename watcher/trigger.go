package watcher

import (
	"errors"
	"log"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type FileHandlerTrigger interface {
	IsTriggeredBy(filePath string) bool
}

type fileListHandlerTrigger struct {
	FileHandlerTrigger

	files []string
}

func newFileListHandlerTrigger(config FileHandlerConfig) (self *fileListHandlerTrigger, err error) {
	if config.Files == nil {
		return nil, errors.New("not a fileListHandlerTrigger config")
	}
	self = &fileListHandlerTrigger{}
	self.files = make([]string, len(config.Files))
	for i, file := range config.Files {
		self.files[i], err = filepath.Abs(file)
		if err != nil {
			return nil, err
		}
	}
	return
}

func (self *fileListHandlerTrigger) IsTriggeredBy(filePath string) bool {
	// Make sure file path is absolute
	if !filepath.IsAbs(filePath) {
		log.Fatalf("malformed file path found by file handler trigger: %s", filePath)
	}
	return slices.Contains(self.files, filePath)
}

type extensionFileHandlerTrigger struct {
	FileHandlerTrigger

	extensions []string
}

func newExtensionFileHandlerTrigger(config FileHandlerConfig) (self *extensionFileHandlerTrigger, err error) {
	if config.Extensions == nil {
		return nil, errors.New("not an extensionFileHandlerTrigger config")
	}
	self = &extensionFileHandlerTrigger{}
	self.extensions = make([]string, len(config.Extensions))
	for i, ext := range config.Extensions {
		if !strings.HasPrefix(ext, ".") {
			ext = strings.Join([]string{".", ext}, "")
		}
		self.extensions[i] = ext
	}
	return
}

func (self *extensionFileHandlerTrigger) IsTriggeredBy(filePath string) bool {
	for _, ext := range self.extensions {
		if strings.HasSuffix(filePath, ext) {
			return true
		}
	}
	return false
}

type regexFileHandlerTrigger struct {
	FileHandlerTrigger

	regex *regexp.Regexp
}

func newRegexFileHandlerTrigger(config FileHandlerConfig) (self *regexFileHandlerTrigger, err error) {
	if config.Regex == "" {
		return nil, errors.New("not a regexFileHandlerTrigger config")
	}
	self = &regexFileHandlerTrigger{}
	self.regex, err = regexp.Compile(config.Regex)
	return
}

func (self *regexFileHandlerTrigger) IsTriggeredBy(filePath string) bool {
	return self.regex.Match([]byte(filePath))
}
