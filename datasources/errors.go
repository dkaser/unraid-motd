package datasources

import (
	"errors"
	"fmt"
)

var ErrCommandFailed = errors.New("Command failed")
var ErrConfigFile = errors.New("Config file error")
var ErrParse = errors.New("Parse error")

func CommandFailedError(message string) error {
	return fmt.Errorf("Command Failed %w : %s", ErrCommandFailed, message)
}

func ConfigFileError(message string) error {
	return fmt.Errorf("Config file error %w : %s", ErrConfigFile, message)
}

func ParseError(message string) error {
	return fmt.Errorf("Parse error %w : %s", ErrParse, message)
}
