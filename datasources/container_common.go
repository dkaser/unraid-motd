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

func (cl *containerList) getContent(ignoreList []string, warnOnly bool, c TableConfig) (content string, err error) {
	t := GetTableWriter(c)
	var title string
	
	// Make set of ignored containers
	var ignoreSet utils.StringSet
	ignoreSet = ignoreSet.FromList(ignoreList)
	// Process output
	var goodCont = make(map[string]string)
	var failedCont = make(map[string]string)
	var sortedNames []string
	for _, c := range cl.Containers {
		if ignoreSet.Contains(c.Name) {
			continue
		}
		status := strings.ToLower(c.Status)
		if status == "up" || status == "created" || status == "running" {
			goodCont[c.Name] = status
		} else {
			failedCont[c.Name] = status
		}
		sortedNames = append(sortedNames, c.Name)
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
			t.AppendRow([]interface{}{c, utils.Good(val)})
		} else if val, ok := failedCont[c]; ok {
			t.AppendRow([]interface{}{c, utils.Err(val)})
		}
	}

	content = RenderTable(t, title)
	return
}
