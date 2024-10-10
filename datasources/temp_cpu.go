package datasources

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/shirou/gopsutil/v3/host"
	log "github.com/sirupsen/logrus"

	"github.com/dkaser/unraid-motd/utils"
)

// ConfTempCPU extends ConfBase with a list of containers to ignore
type ConfTempCPU struct {
	ConfBaseWarn `yaml:",inline"`
}

// Init sets up default alignment
func (c *ConfTempCPU) Init() {
	c.ConfBaseWarn.Init()
}

// GetCPUTemp returns CPU core temps using gopsutil or parsing sensors output
func GetCPUTemp(channel chan<- SourceReturn, conf *Conf) {
	sourceConf := conf.CPU
	sourceConf.Load(conf)

	returnData := NewSourceReturn(conf.debug)
	defer func() {
		channel <- returnData.Return()
	}()
	var tempMap map[string]int
	var isZen bool
	var err error
	tempMap, isZen, err = cpuTempGopsutil()

	if err != nil {
		log.Warnf("[cpu] temperature read error: %v", err)
	}

	if len(tempMap) == 0 {
		outputTable := GetTableWriter(sourceConf)
		returnData.Content = RenderTable(outputTable, "CPU Temp: "+utils.Warn("Unavailable"))
	} else {
		returnData.Content = formatCPUTemps(tempMap, isZen, &sourceConf)
	}
}

func formatCPUTemps(tempMap map[string]int, isZen bool, sourceConf *ConfTempCPU) (content string) {
	outputTable := GetTableWriter(sourceConf)
	var title string

	// Sort keys
	sortedNames := make([]string, len(tempMap))
	i := 0
	for core := range tempMap {
		sortedNames[i] = core
		i++
	}
	sort.Strings(sortedNames)
	var warnCount int
	var errCount int
	for _, core := range sortedNames {
		coreTemp := tempMap[core]
		var wrapped string
		if !isZen {
			wrapped = fmt.Sprintf("Core %s", core)
		} else {
			wrapped = core
		}
		if coreTemp < sourceConf.Warn && !*sourceConf.WarnOnly {
			outputTable.AppendRow([]interface{}{wrapped, utils.Good(coreTemp)})
		} else if coreTemp >= sourceConf.Warn && coreTemp < sourceConf.Crit {
			outputTable.AppendRow([]interface{}{wrapped, utils.Warn(coreTemp)})
			warnCount++
		} else if coreTemp >= sourceConf.Crit {
			warnCount++
			errCount++
			outputTable.AppendRow([]interface{}{wrapped, utils.Err(coreTemp)})
		}
	}
	if warnCount == 0 {
		title = fmt.Sprintf("%s: %s", "CPU Temp", utils.Good("OK"))
	} else if errCount > 0 {
		title = fmt.Sprintf("%s: %s", "CPU Temp", utils.Err("Critical"))
	} else if warnCount > 0 {
		title = fmt.Sprintf("%s: %s", "CPU Temp", utils.Warn("Warning"))
	}

	content = RenderTable(outputTable, title)

	return
}

func cpuTempGopsutil() (tempMap map[string]int, isZen bool, err error) {
	temps, err := host.SensorsTemperatures()
	tempMap = make(map[string]int)
	addTemp := func(re *regexp.Regexp) {
		for _, stat := range temps {
			log.Debugf("[cpu] check %s", stat.SensorKey)
			m := re.FindStringSubmatch(stat.SensorKey)
			if len(m) > 1 {
				log.Debugf("[cpu] OK %s: %.0f", stat.SensorKey, stat.Temperature)
				tempMap[m[1]] = int(stat.Temperature)
			}
		}
	}
	addTemp(regexp.MustCompile(`coretemp_core(?:_)?(\d+)`))
	// Try k10temp if we didn't find anything
	if len(tempMap) == 0 {
		isZen = true
		log.Debug("[cpu] trying k10temp")
		addTemp(regexp.MustCompile(`k10temp_(\w+)`))
	}
	// Something's really wrong if we still have none
	if len(tempMap) == 0 {
		log.Warn("[cpu] could not find any CPU temperatures")
	} else {
		err = nil
	}

	return
}
