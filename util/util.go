package util

import (
	"fmt"
	"os"
)

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

func WrapError(msg string, err error) error {
	return fmt.Errorf("%s\n\t%w", msg, err)
}
