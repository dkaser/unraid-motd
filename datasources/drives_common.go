package datasources

import (
	"fmt"
	"strings"

	"github.com/dkaser/unraid-motd/utils"
	zfs "github.com/mistifyio/go-zfs/v3"
	"github.com/shirou/gopsutil/v3/disk"
)

type ConfDrives struct {
	ConfBaseWarn    `yaml:",inline"`
	ShowZFSDatasets bool `yaml:"show_zfs_datasets"`
}

// Init is mandatory
func (c *ConfDrives) Init() {
	// Base init must be called
	c.ConfBaseWarn.Init()
	c.WarnOnly = new(bool)
	c.ShowZFSDatasets = false
}

func getSystemDirs() []string {
	return []string{"/var/log", "/boot", "/var/lib/docker"}
}

func processDrive(sourceConf *ConfDrives, mountpoint string, fsType string, status string) (newStatus string, percent int, used string, total string, err error) {
	switch fsType {
	case "zfs":
		newStatus, percent, used, total, err = processDriveZFS(sourceConf, mountpoint, status)
	default:
		newStatus, percent, used, total = processDriveGeneric(sourceConf, mountpoint, status)
	}

	return
}

func getStatus(sourceConf *ConfDrives, status string, percent int) string {
	if percent >= sourceConf.Warn && percent < sourceConf.Crit {
		if status != "e" {
			status = "w"
		}
	} else if percent >= sourceConf.Crit {
		status = "e"
	}

	return status
}

func processDriveZFS(sourceConf *ConfDrives, mountpoint string, status string) (newStatus string, percent int, used string, total string, err error) {
	dataset, _ := zfs.GetDataset(mountpoint)

	if (!sourceConf.ShowZFSDatasets) && strings.Contains(dataset.Name, "/") {
		err = ZFSError("Skipping dataset")

		return
	}

	usedVal := float64(dataset.Used)
	totalVal := float64(dataset.Used) + float64(dataset.Avail)

	used = utils.FormatBytes(usedVal)
	total = utils.FormatBytes(totalVal)
	percent = int((usedVal / totalVal) * 100)

	newStatus = getStatus(sourceConf, status, percent)

	return
}

func processDriveGeneric(sourceConf *ConfDrives, mountpoint string, status string) (newStatus string, percent int, used string, total string) {
	diskUsage, _ := disk.Usage(mountpoint)

	used = utils.FormatBytes(float64(diskUsage.Used))
	total = utils.FormatBytes(float64(diskUsage.Total))
	percent = int(diskUsage.UsedPercent)

	newStatus = getStatus(sourceConf, status, percent)

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
