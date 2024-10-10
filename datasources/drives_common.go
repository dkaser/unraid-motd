package datasources

import (
	"fmt"
	"github.com/dkaser/unraid-motd/utils"
	"github.com/shirou/gopsutil/v3/disk"
)

type ConfDrives struct {
	ConfBaseWarn `yaml:",inline"`
}

// Init is mandatory
func (c *ConfDrives) Init() {
	// Base init must be called
	c.ConfBaseWarn.Init()
	c.WarnOnly = new(bool)
}

func getSystemDirs() []string {
	return []string{"/var/log", "/boot", "/var/lib/docker"}
}

func processDrive(sourceConf *ConfDrives, mountpoint string, status string) (newStatus string, percent int, used string, total string) {
	diskUsage, _ := disk.Usage(mountpoint)

	used = utils.FormatBytes(float64(diskUsage.Used))
	total = utils.FormatBytes(float64(diskUsage.Total))
	percent = int(diskUsage.UsedPercent)

	if percent >= sourceConf.Warn && percent < sourceConf.Crit {
		if status != "e" {
			status = "w"
		}
	} else if percent >= sourceConf.Crit {
		status = "e"
	}

	newStatus = status

	return
}

func formatDriveUsage(sourceConf *ConfDrives, percent int) string {
	text := fmt.Sprintf("%d%%", percent)

	if percent >= sourceConf.Warn && percent < sourceConf.Crit {
		return utils.Warn(text)
	} else if percent >= sourceConf.Crit {
		return utils.Err(text)
	}

	return utils.Good(text)
}

func getDriveHeaderTable(title string, status string) (header string) {
	if status == "o" {
		header = fmt.Sprintf("%s: %s", title, utils.Good("OK"))
	} else if status == "w" {
		header = fmt.Sprintf("%s: %s", title, utils.Warn("Warning"))
	} else {
		header = fmt.Sprintf("%s: %s", title, utils.Err("Critical"))
	}

	return
}
