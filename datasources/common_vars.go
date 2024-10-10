package datasources

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type UnavailableError interface {
	error
	UnavailableError()
}

type ModuleNotAvailable struct {
	Name        string
	ParentError error
}

func (m *ModuleNotAvailable) Error() string {
	return "module " + m.Name + " is not available: " + m.ParentError.Error()
}

func (ModuleNotAvailable) UnavailableError() {}

// SourceReturn is the data returned by a datasource through a channel
type SourceReturn struct {
	// Datasource output content string
	Content string
	// Error
	Error error
	// Time taken, non-zero only in debug mode
	Time time.Duration
	// Internal
	start time.Time
}

func (sr *SourceReturn) Return() SourceReturn {
	if !sr.start.IsZero() {
		sr.Time = time.Since(sr.start)
	}

	return *sr
}

func NewSourceReturn(debug bool) *SourceReturn {
	sr := SourceReturn{}
	if debug {
		sr.start = time.Now()
	}

	return &sr
}

// ConfInterface defines the interface for config structs
type ConfInterface interface {
	Init()
}

func NewConfFromFile(path string, debug bool) (c Conf, err error) {
	c.Init()
	c.debug = debug
	yamlFile, errF := os.ReadFile(path)
	if errors.Is(errF, fs.ErrNotExist) {
		return
	}
	if errF != nil {
		err = ConfigFileError(errF.Error())

		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		err = ParseError(fmt.Sprintf("%s: %v", path, err))

		return
	}

	return
}

// RunSources runs data sources in runList, the names are validated and returned as the first value
func RunSources(runList []string, conf *Conf) ([]string, map[string]SourceReturn) {
	channels := make(map[string]chan SourceReturn)
	out := make(map[string]SourceReturn)
	var validRuns []string
	// Start goroutines
Loop:
	for _, source := range runList {
		channel := make(chan SourceReturn, 1)
		switch source {
		case "cpu":
			go GetCPUTemp(channel, conf)
		case "docker":
			go GetDocker(channel, conf)
		case "sysinfo":
			go GetSysInfo(channel, conf)
		case "user-drives":
			go GetUserDrives(channel, conf)
		case "system-drives":
			go GetSystemDrives(channel, conf)
		case "networks":
			go GetNetworks(channel, conf)
		case "services":
			go GetServices(channel, conf)
		default:
			log.Warnf("no data source named %s", source)

			continue Loop
		}
		channels[source] = channel
		validRuns = append(validRuns, source)
	}
	// Wait for results
	log.Debug("Wait for goroutines")
	for source := range channels {
		out[source] = <-channels[source]
	}

	return validRuns, out
}
