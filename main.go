package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"gopkg.in/yaml.v2"

	"github.com/dkaser/unraid-motd/datasources"

	"github.com/arsham/figurine/figurine"
	"golang.org/x/term"
)

var defaultCfgPath = "./config.yaml"

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
	// Flatten 2-dim input
	for _, row := range c.ColDef {
		printOrder = append(printOrder, row...)
	}

	return
}

var args struct {
	ConfigFile      string `arg:"-c,--config,env:CONFIG_FILE"             help:"Path to config yaml"`
	Debug           bool   `arg:"--debug,env:DEBUG"                       help:"Debug mode"`
	DumpConfig      bool   `arg:"--dump-config"                           help:"Dump config and exit"`
	HideUnavailable bool   `arg:"--hide-unavailable,env:HIDE_UNAVAILABLE" help:"Hide unavailable modules"`
	LogLevel        string `arg:"--log-level,env:LOG_LEVEL"               help:"Set log level"`
	PID             string `arg:"--pid"                                   help:"Write PID to file or log if '-'"`
	Quiet           bool   `arg:"-q,--quiet"                              help:"Don't log to console"`
}

func setupLogging() {
	var logLevel log.Level
	defaultLevel := log.WarnLevel

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

func runModules(conf *datasources.Conf) {
	outOrder, outData := datasources.RunSources(makePrintOrder(conf), conf)
	outStr := make(map[string]string)
	// Wait and save results
	for _, source := range outOrder {
		sourceData, ok := outData[source]
		if !ok {
			continue
		}
		// Check if we should skip due to unavailable error
		if _, unOK := sourceData.Error.(datasources.UnavailableError); unOK && args.HideUnavailable {
			continue
		}
		if sourceData.Error != nil {
			log.Warnf("%s error: %v", source, sourceData.Error)
		}

		if sourceData.Content != "" {
			outStr[source] = sourceData.Content
		}
	}
	outBuf := &strings.Builder{}

	mapToTable(outBuf, outStr, conf.ColDef, conf.ColPad)

	fmt.Print(outBuf.String())

	// Show timing results
	if args.Debug {
		for _, k := range outOrder {
			log.Debugf("%s ran in: %s", k, outData[k].Time.String())
		}
	}
}

func main() {
	if !term.IsTerminal(0) {
		return
	}
	width, _, err := term.GetSize(0)
	if err != nil {
		return
	}

	args.ConfigFile = defaultCfgPath
	arg.MustParse(&args)

	setupLogging()

	var mainStart time.Time
	if args.Debug {
		mainStart = time.Now()
	}
	// Read config file
	conf, err := datasources.NewConfFromFile(args.ConfigFile, args.Debug)
	if err != nil {
		log.Warn(err)
	}

	if conf.FixedTableWidth > width {
		conf.FixedTableWidth = width
	}

	if args.DumpConfig {
		log.Info("Dumping config")
		if flag.NArg() > 0 {
			dumpConfig(&conf, flag.Arg(0))
		} else {
			dumpConfig(&conf, "")
		}

		return
	}

	if conf.Header.Show {
		text := conf.Header.CustomText
		if conf.Header.UseHostname {
			text, _ = os.Hostname()
		}

		err := figurine.Write(os.Stdout, text, conf.Header.Font)
		if err != nil {
			log.Debug(err.Error())
		}

		fmt.Println("")
	}

	runModules(&conf)

	// Show timing results
	if args.Debug {
		log.Debugf("main ran in: %s", time.Since(mainStart).String())
	}
}

func dumpConfig(c *datasources.Conf, writeFile string) {
	configDump, err := yaml.Marshal(c)
	if err != nil {
		log.Errorf("Config parse error: %v", err)

		return
	}
	if writeFile != "" {
		err = os.WriteFile(writeFile, configDump, 0600)
		if err != nil {
			log.Errorf("Config dumped failed: %v", err)

			return
		}
		log.Infof("Config dumped to: %s", writeFile)
	} else {
		fmt.Printf("%s\n", string(configDump))
	}
}
