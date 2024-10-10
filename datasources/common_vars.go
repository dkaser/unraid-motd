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
		err = fmt.Errorf("config file error: %v ", errF)

		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		err = fmt.Errorf("cannot parse %s: %v", path, err)

		return
	}

	return
}

// RunSources runs data sources in runList, the names are validated and returned as the first value
func RunSources(runList []string, c *Conf) ([]string, map[string]SourceReturn) {
	channels := make(map[string]chan SourceReturn)
	out := make(map[string]SourceReturn)
	var validRuns []string
	// Start goroutines
Loop:
	for _, k := range runList {
		ch := make(chan SourceReturn, 1)
		switch k {
		case "cpu":
			go GetCPUTemp(ch, c)
		case "docker":
			go GetDocker(ch, c)
		case "sysinfo":
			go GetSysInfo(ch, c)
		case "user-drives":
			go GetUserDrives(ch, c)
		case "system-drives":
			go GetSystemDrives(ch, c)
		case "networks":
			go GetNetworks(ch, c)
		case "services":
			go GetServices(ch, c)
		default:
			log.Warnf("no data source named %s", k)

			continue Loop
		}
		channels[k] = ch
		validRuns = append(validRuns, k)
	}
	// Wait for results
	log.Debug("Wait for goroutines")
	for k := range channels {
		out[k] = <-channels[k]
	}

	return validRuns, out
}
