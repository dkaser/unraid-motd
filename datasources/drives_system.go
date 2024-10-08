package datasources

import (
  "github.com/shirou/gopsutil/v3/disk"
  "slices"
  "fmt"
  "github.com/jedib0t/go-pretty/v6/table"
  "github.com/jedib0t/go-pretty/v6/text"
)

func GetSystemDrives(ch chan<- SourceReturn, conf *Conf) {
	c := conf.SystemDrives
	sr := NewSourceReturn(conf.debug)
	defer func() {
		ch <- sr.Return(&c.ConfBase)
	}()
    sr.Header, sr.Content, sr.Error = getSystemDriveUsage(&c)
    return
}

func getSystemDriveUsage(c *ConfDrives) (header string, content string, err error) {
	status := "o"

	parts, err := disk.Partitions(true)
	if err != nil {
		err = &ModuleNotAvailable{"drives", err}
		return
	}

	t := table.NewWriter()

	PARTITIONS:
	for _, p := range parts {
		if(!slices.Contains(getSystemDirs(), p.Mountpoint)) {
			continue PARTITIONS
		}	

		newStatus, percent, used, total, _ := processDrive(c,p.Mountpoint,status,content)
		status = newStatus
		t.AppendRow([]interface{}{p.Mountpoint, formatDriveUsage(c, percent), fmt.Sprintf("%s %s %s", used, "used out of", total)})
	}

	title, _ := getDriveHeaderTable(c, "System Drives", status)

	t.SetStyle(GetTableStyle())
	t.SetColumnConfigs([]table.ColumnConfig{
        {Number: 1, Align: text.AlignRight},
    })
	t.Style().Options.SeparateRows = false

	t.SetTitle(title)

	content = t.Render();

	
	return
}