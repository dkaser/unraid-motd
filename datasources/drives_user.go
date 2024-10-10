package datasources

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
	"slices"
	"strings"
)

func GetUserDrives(ch chan<- SourceReturn, conf *Conf) {
	c := conf.UserDrives
	c.Load(conf)

	sr := NewSourceReturn(conf.debug)
	defer func() {
		ch <- sr.Return()
	}()
	sr.Content, sr.Error = getUserDriveUsage(&c)
}

func getUserDriveUsage(c *ConfDrives) (content string, err error) {
	t := GetTableWriter(c)

	deviceIgnore := []string{"docker/", "shfs", "nfsd", "nsfs", "/loop"}
	allowedFs := []string{"vfat", "xfs", "btrfs", "zfs"}
	status := "o"

	parts, err := disk.Partitions(true)
	if err != nil {
		err = &ModuleNotAvailable{"drives", err}

		return
	}

PARTITIONS:
	for _, p := range parts {
		if !slices.Contains(allowedFs, p.Fstype) {
			continue
		}

		for _, s := range deviceIgnore {
			if strings.Contains(p.Device, s) {
				continue PARTITIONS
			}
		}

		if slices.Contains(getSystemDirs(), p.Mountpoint) {
			continue PARTITIONS
		}

		newStatus, percent, used, total := processDrive(c, p.Mountpoint, status)
		status = newStatus

		if (percent >= c.Warn) || (!*c.WarnOnly) {
			t.AppendRow([]interface{}{p.Mountpoint, formatDriveUsage(c, percent), fmt.Sprintf("%s %s %s", used, "used out of", total)})
		}
	}

	content = RenderTable(t, getDriveHeaderTable("User Drives", status))

	return
}
