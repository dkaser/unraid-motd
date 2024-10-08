package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"gopkg.in/yaml.v2"

	"github.com/cosandr/go-motd/datasources"
	"github.com/cosandr/go-motd/utils"
)

const defaultRefresh string = "10m"

var defaultCfgPath = "./config.yaml"
var defaultOrder = []string{"sysinfo", "cpu", "docker", "user-drives", "system-drives", "networks"}

func makeTable(buf *strings.Builder, padding int) (table *tablewriter.Table) {
	table = tablewriter.NewWriter(buf)
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding(strings.Repeat(" ", padding))
	table.SetNoWhiteSpace(true)
	return
}

func mapToTable(buf *strings.Builder, inStr map[string]string, colDef [][]string, padding int) {
	table := makeTable(buf, padding)
	var tmp []string
	// Render a new table every row for compact output
	for _, row := range colDef {
		// Just write block to buffer if it is alone
		if len(row) == 1 {
			a, ok := inStr[row[0]]
			// Skip invalid modules
			if !ok {
				continue
			}
			_, _ = fmt.Fprintln(buf, a)
			continue
		}
		tmp = nil
		for _, k := range row {
			a, ok := inStr[k]
			if !ok {
				continue
			}
			tmp = append(tmp, a)
		}
		table.Append(tmp)
		table.Render()
		// Remake table to avoid imbalanced output
		table = makeTable(buf, padding)
	}
}

// makePrintOrder flattens colDef (if present). If showOrder is defined as well, it is ignored.
func makePrintOrder(c *datasources.Conf) (printOrder []string) {
	if len(c.ColDef) > 0 {
		// Flatten 2-dim input
		for _, row := range c.ColDef {
			for _, k := range row {
				printOrder = append(printOrder, k)
			}
		}
	} else if len(c.ShowOrder) > 0 {
		printOrder = c.ShowOrder
	} else {
		// Use default order
		printOrder = make([]string, len(defaultOrder))
		copy(printOrder, defaultOrder)
	}
	return
}

var args struct {
	ConfigFile      string        `arg:"-c,--config,env:CONFIG_FILE" help:"Path to config yaml"`
	Daemon          bool          `arg:"-d,--daemon,env:DAEMON" help:"Run in daemon mode"`
	Debug           bool          `arg:"--debug,env:DEBUG" help:"Debug mode"`
	DumpConfig      bool          `arg:"--dump-config" help:"Dump config and exit"`
	HideUnavailable bool          `arg:"--hide-unavailable,env:HIDE_UNAVAILABLE" help:"Hide unavailable modules"`
	LogLevel        string        `arg:"--log-level,env:LOG_LEVEL" help:"Set log level"`
	NoColors        bool          `arg:"--no-colors,env:NO_COLORS" help:"Disable colors"`
	Output          string        `arg:"-o,--output,env:OUTPUT" help:"Write output to file instead of stdout"`
	PID             string        `arg:"--pid" help:"Write PID to file or log if '-'"`
	Quiet           bool          `arg:"-q,--quiet" help:"Don't log to console"`
	RefreshInterval time.Duration `arg:"--refresh-interval,env:REFRESH_INTERVAL" help:"Time interval between data refreshes"`
}

func setupLogging() {
	var logLevel log.Level
	var defaultLevel log.Level
	if args.Daemon {
		defaultLevel = log.InfoLevel
	} else {
		defaultLevel = log.WarnLevel
	}
	var err error
	getLogLevels := func(level log.Level) []log.Level {
		ret := make([]log.Level, 0)
		for _, lvl := range log.AllLevels {
			if level >= lvl {
				ret = append(ret, lvl)
			}
		}
		return ret
	}

	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	log.SetOutput(io.Discard)
	if args.Debug {
		logLevel = log.DebugLevel
	} else if args.LogLevel != "" {
		logLevel, err = log.ParseLevel(args.LogLevel)
		if err != nil {
			logLevel = defaultLevel
			log.Warnf("Unknown log level %s, defaulting to %s", args.LogLevel, logLevel.String())
		}
	} else {
		logLevel = defaultLevel
	}
	log.SetLevel(logLevel)
	levels := getLogLevels(logLevel)
	if !args.Quiet {
		log.AddHook(&writer.Hook{
			Writer:    os.Stderr,
			LogLevels: levels,
		})
	}
}

func runModules(c *datasources.Conf) {
	outOrder, outData := datasources.RunSources(makePrintOrder(c), c)
	outStr := make(map[string]string)
	// Wait and save results
	for _, k := range outOrder {
		v, ok := outData[k]
		if !ok {
			continue
		}
		// Check if we should skip due to unavailable error
		if _, unOK := v.Error.(datasources.UnavailableError); unOK && args.HideUnavailable {
			continue
		}
		if v.Error != nil {
			log.Warnf("%s error: %v", k, v.Error)
		}
		outStr[k] = v.Header
		if v.Content != "" {
			outStr[k] += "\n" + v.Content
		}
	}
	outBuf := &strings.Builder{}
	if len(c.ColDef) > 0 {
		log.Debug("Format as table")
		mapToTable(outBuf, outStr, c.ColDef, c.ColPad)
	} else {
		log.Debug("Print as is")
		for _, k := range outOrder {
			_, _ = fmt.Fprintln(outBuf, outStr[k])
		}
	}
	if args.Output != "" {
		err := os.WriteFile(args.Output, []byte(outBuf.String()), 0644)
		if err != nil {
			log.Error(err)
		}
	} else {
		fmt.Print(outBuf.String())
	}
	// Show timing results
	if args.Debug {
		for _, k := range outOrder {
			log.Debugf("%s ran in: %s", k, outData[k].Time.String())
		}
	}
}

func runDaemon(c *datasources.Conf) {
	if args.PID == "-" {
		log.Infof("PID: %d", os.Getpid())
	} else if args.PID != "" {
		err := os.WriteFile(args.PID, []byte(fmt.Sprint(os.Getpid())), 0644)
		if err != nil {
			log.Errorf("cannot write PID: %v", err)
		}
	}
	defer func() {
		// Delete PID file if it exists
		if args.PID != "-" && args.PID != "" {
			info, err := os.Stat(args.PID)
			if os.IsNotExist(err) {
				return
			}
			if !info.IsDir() {
				if err := os.Remove(args.PID); err != nil {
					log.Error(err)
				}
			}
		}
	}()
	log.Infof("auto-refresh every %v", args.RefreshInterval)
	var refreshStart time.Time
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	ticker := time.NewTicker(args.RefreshInterval)
	// Always run at startup
	runModules(c)
	for {
		select {
		case <-ticker.C:
			log.Debug("auto-refresh ticker")
			if args.Debug {
				refreshStart = time.Now()
			}
			runModules(c)
			if args.Debug {
				log.Debugf("refresh ran in: %s", time.Now().Sub(refreshStart).String())
			}
		case s := <-signals:
			switch s {
			case syscall.SIGHUP:
				log.Debug("SIGHUP received, refreshing")
				runModules(c)
				ticker.Reset(args.RefreshInterval)
			default:
				log.Warn("exit signal received")
				return
			}
		}
	}
}

func main() {
	args.ConfigFile = defaultCfgPath
	args.RefreshInterval, _ = time.ParseDuration(defaultRefresh)
	arg.MustParse(&args)

	setupLogging()

	var mainStart time.Time
	if args.Debug {
		mainStart = time.Now()
	}
	if args.NoColors {
		utils.NoColors = true
	}
	// Read config file
	c, err := datasources.NewConfFromFile(args.ConfigFile, args.Debug)
	if err != nil {
		log.Warn(err)
	}

	if args.DumpConfig {
		log.Info("Dumping config")
		if flag.NArg() > 0 {
			dumpConfig(&c, flag.Arg(0))
		} else {
			dumpConfig(&c, "")
		}
		return
	}

	if args.Daemon {
		runDaemon(&c)
	} else {
		runModules(&c)
	}
	// Show timing results
	if args.Debug {
		log.Debugf("main ran in: %s", time.Now().Sub(mainStart).String())
	}
}

func dumpConfig(c *datasources.Conf, writeFile string) {
	d, err := yaml.Marshal(c)
	if err != nil {
		log.Errorf("Config parse error: %v", err)
		return
	}
	if writeFile != "" {
		err = os.WriteFile(writeFile, d, 0644)
		if err != nil {
			log.Errorf("Config dumped failed: %v", err)
			return
		}
		log.Infof("Config dumped to: %s", writeFile)
	} else {
		fmt.Printf("%s\n", string(d))
	}
	return
}
