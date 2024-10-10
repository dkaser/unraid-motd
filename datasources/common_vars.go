package datasources

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"
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

func (sr *SourceReturn) Return(c *ConfBase) SourceReturn {
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

// ConfBase is the common type for all modules
//
// Custom modules should respect these options
type ConfBase struct {
	// Override global setting
	WarnOnly *bool `yaml:"warnings_only,omitempty"`
	FixedTableWidth *int `yaml:"table_width,omitempty"`
}

func (c *ConfBase) Init() {

}

// ConfBaseWarn extends ConfBase with warning and critical values
type ConfBaseWarn struct {
	ConfBase `yaml:",inline"`
	Warn     int `yaml:"warn"`
	Crit     int `yaml:"crit"`
}

// Init sets warning to 70 and critical to 90
func (c *ConfBaseWarn) Init() {
	c.ConfBase.Init()
	c.Warn = 70
	c.Crit = 90
}

// ConfGlobal is the config struct for global settings
type ConfGlobal struct {
	// Hide fields which are deemed to be OK
	WarnOnly bool `yaml:"warnings_only"`
	// Order in which to display data sources
	ShowOrder []string `yaml:"show_order,flow,omitempty"`
	// Define how data sources are displayed
	ColDef [][]string `yaml:"col_def,flow,omitempty"`
	// Padding between columns when using col_def
	ColPad int `yaml:"col_pad"`
	// Internal variables
	debug bool

	FixedTableWidth int `yaml:"table_width"`
}

type ConfHeader struct {
	Show bool `yaml:"show"`
	UseHostname bool `yaml:"use_hostname"`
	CustomText string `yaml:"custom_text"`
	Font string `yaml:"font"`
}

func (c *ConfHeader) Init() {
	c.Show = true
	c.UseHostname = true
	c.CustomText = "Custom"
	c.Font = "Banner.flf"
}

// Conf is the combined config struct, defines YAML file
type Conf struct {
	ConfGlobal   `yaml:"global"`
	Header		 ConfHeader   `yaml:"header"`
	CPU          ConfTempCPU  `yaml:"cpu"`
	Docker       ConfDocker   `yaml:"docker"`
	SysInfo      ConfSysInfo  `yaml:"sysinfo"`
	UserDrives   ConfDrives   `yaml:"user-drives"`
	SystemDrives ConfDrives   `yaml:"system-drives"`
	Networks     ConfNet      `yaml:"network"`
	Services     ConfServices `yaml:"services"`
}

// Init a config with sane default values
func (c *Conf) Init() {
	// Set global defaults
	c.WarnOnly = true
	c.ColPad = 4
	c.FixedTableWidth = 0
	c.Header.Init()
	// Init data source configs
	c.CPU.Init()
	c.Docker.Init()
	c.SysInfo.Init()
	c.UserDrives.Init()
	c.SystemDrives.Init()
	c.Networks.Init()
	c.Services.Init()
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

type timeEntry struct {
	short string
	long  string
}

// timeStr returns human friendly time durations
func timeStr(d time.Duration, precision int, short bool) string {
	times := map[int]timeEntry{
		1:            {"s", "second"},
		60:           {"m", "minute"},
		3600:         {"h", "hour"},
		86400:        {"d", "day"},
		604800:       {"w", "week"},
		int(2.628e6): {"mo", "month"},
		int(3.154e7): {"yr", "year"},
	}
	// Sort keys to ensure proper order
	keys := make([]int, 0)
	for k := range times {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	seconds := int(d.Seconds())
	if seconds < 1 {
		return "just now"
	}
	var ret string
	var tmp int
	for _, k := range keys {
		if tmp >= precision {
			break
		}
		q := seconds / k
		r := seconds % k
		// We have <1 of this unit
		if q == 0 {
			continue
		}
		if short {
			ret += fmt.Sprintf("%d%s", q, times[k].short)
		} else {
			if q == 1 {
				// We have one, don't add s
				ret += fmt.Sprintf("%d %s, ", q, times[k].long)
			} else {
				// More than one or zero, add s at the end
				ret += fmt.Sprintf("%d %ss, ", q, times[k].long)
			}
		}
		seconds = r
		tmp++
	}
	return strings.TrimSuffix(ret, ", ")
}
