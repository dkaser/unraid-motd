package datasources

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dkaser/unraid-motd/utils"
)

type containerStatus struct {
	Name   string
	Status string
}

type containerList struct {
	Runtime    string
	Root       bool
	Containers []containerStatus
}

func (cl *containerList) getContent(ignoreList []string, warnOnly bool, sourceConf TableConfig) (content string) {
	outputTable := GetTableWriter(sourceConf)
	var title string

	// Make set of ignored containers
	var ignoreSet utils.StringSet
	ignoreSet = ignoreSet.FromList(ignoreList)
	// Process output
	var goodCont = make(map[string]string)
	var failedCont = make(map[string]string)
	var sortedNames []string
	for _, container := range cl.Containers {
		if ignoreSet.Contains(container.Name) {
			continue
		}
		status := strings.ToLower(container.Status)
		if status == "up" || status == "created" || status == "running" {
			goodCont[container.Name] = status
		} else {
			failedCont[container.Name] = status
		}
		sortedNames = append(sortedNames, container.Name)
	}
	sort.Strings(sortedNames)

	// Decide what header should be
	if len(goodCont) == 0 && len(sortedNames) > 0 {
		title = fmt.Sprintf("%s: %s", cl.Runtime, utils.Err("Critical"))
	} else if len(failedCont) == 0 {
		title = fmt.Sprintf("%s: %s", cl.Runtime, utils.Good("OK"))
	} else if len(failedCont) < len(sortedNames) {
		title = fmt.Sprintf("%s: %s", cl.Runtime, utils.Warn("Warning"))
	}

	// Only print all containers if requested
	for _, c := range sortedNames {
		if val, ok := goodCont[c]; ok && !warnOnly {
			outputTable.AppendRow([]interface{}{c, utils.Good(val)})
		} else if val, ok := failedCont[c]; ok {
			outputTable.AppendRow([]interface{}{c, utils.Err(val)})
		}
	}

	content = RenderTable(outputTable, title)

	return
}
