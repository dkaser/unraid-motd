package datasources

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
	"slices"
)

func GetSystemDrives(ch chan<- SourceReturn, conf *Conf) {
	c := conf.SystemDrives
	c.Load(conf)

	sr := NewSourceReturn(conf.debug)
	defer func() {
		ch <- sr.Return()
	}()
	sr.Content, sr.Error = getSystemDriveUsage(&c)
}

func getSystemDriveUsage(c *ConfDrives) (content string, err error) {
	t := GetTableWriter(*c)

	status := "o"

	parts, err := disk.Partitions(true)
	if err != nil {
		err = &ModuleNotAvailable{"drives", err}

		return
	}

PARTITIONS:
	for _, p := range parts {
		if !slices.Contains(getSystemDirs(), p.Mountpoint) {
			continue PARTITIONS
		}

		newStatus, percent, used, total := processDrive(c, p.Mountpoint, status)
		status = newStatus
		if (percent >= c.Warn) || (!*c.WarnOnly) {
			t.AppendRow([]interface{}{p.Mountpoint, formatDriveUsage(c, percent), fmt.Sprintf("%s %s %s", used, "used out of", total)})
		}
	}

	content = RenderTable(t, getDriveHeaderTable("System Drives", status))

	return
}
