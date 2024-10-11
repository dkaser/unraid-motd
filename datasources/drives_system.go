package datasources

import (
	"fmt"
	"slices"

	"github.com/shirou/gopsutil/v3/disk"
)

func GetSystemDrives(channel chan<- SourceReturn, conf *Conf) {
	sourceConf := conf.SystemDrives
	sourceConf.Load(conf)

	returnData := NewSourceReturn(conf.debug)
	defer func() {
		channel <- returnData.Return()
	}()
	returnData.Content, returnData.Error = getSystemDriveUsage(&sourceConf)
}

func getSystemDriveUsage(sourceConf *ConfDrives) (content string, err error) {
	outputTable := GetTableWriter(*sourceConf)

	status := "o"

	parts, err := disk.Partitions(true)
	if err != nil {
		err = &ModuleNotAvailable{"drives", err}

		return
	}

PARTITIONS:
	for _, partition := range parts {
		if !slices.Contains(getSystemDirs(), partition.Mountpoint) {
			continue PARTITIONS
		}

		newStatus, percent, used, total := processDrive(sourceConf, partition.Mountpoint, status)
		status = newStatus
		if (percent >= sourceConf.Warn) || (!*sourceConf.WarnOnly) {
			outputTable.AppendRow([]interface{}{partition.Mountpoint, formatDriveUsage(sourceConf, percent), fmt.Sprintf("%s %s %s", used, "used out of", total)})
		}
	}

	content = RenderTable(outputTable, getDriveHeaderTable("System Drives", status))

	return
}
