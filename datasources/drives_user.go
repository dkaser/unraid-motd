package datasources

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
	"slices"
	"strings"
)

func GetUserDrives(channel chan<- SourceReturn, conf *Conf) {
	sourceConf := conf.UserDrives
	sourceConf.Load(conf)

	sr := NewSourceReturn(conf.debug)
	defer func() {
		channel <- sr.Return()
	}()
	sr.Content, sr.Error = getUserDriveUsage(&sourceConf)
}

func getUserDriveUsage(sourceConf *ConfDrives) (content string, err error) {
	outputTable := GetTableWriter(sourceConf)

	deviceIgnore := []string{"docker/", "shfs", "nfsd", "nsfs", "/loop"}
	allowedFs := []string{"vfat", "xfs", "btrfs", "zfs"}
	status := "o"

	parts, err := disk.Partitions(true)
	if err != nil {
		err = &ModuleNotAvailable{"drives", err}

		return
	}

PARTITIONS:
	for _, partition := range parts {
		if !slices.Contains(allowedFs, partition.Fstype) {
			continue
		}

		for _, s := range deviceIgnore {
			if strings.Contains(partition.Device, s) {
				continue PARTITIONS
			}
		}

		if slices.Contains(getSystemDirs(), partition.Mountpoint) {
			continue PARTITIONS
		}

		newStatus, percent, used, total := processDrive(sourceConf, partition.Mountpoint, status)
		status = newStatus

		if (percent >= sourceConf.Warn) || (!*sourceConf.WarnOnly) {
			outputTable.AppendRow([]interface{}{partition.Mountpoint, formatDriveUsage(sourceConf, percent), fmt.Sprintf("%s %s %s", used, "used out of", total)})
		}
	}

	content = RenderTable(outputTable, getDriveHeaderTable("User Drives", status))

	return
}
