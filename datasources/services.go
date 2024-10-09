package datasources

import (
	"github.com/cosandr/go-motd/utils"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"os/exec"
)

type ConfServices struct {
	ConfBase `yaml:",inline"`
	Services []string `yaml:"pad_header,flow"`
}

// Init is mandatory
func (c *ConfServices) Init() {
	// Base init must be called
	c.ConfBase.Init()
	c.Services = []string{
		"nginx",
		"samba",
		"tailscale",
		"nfsd",
		"sshd",
		"docker",
	}
}

func GetServices(ch chan<- SourceReturn, conf *Conf) {
	c := conf.Services
	if c.FixedTableWidth == nil {
		c.FixedTableWidth = &conf.FixedTableWidth
	}

	sr := NewSourceReturn(conf.debug)
	defer func() {
		ch <- sr.Return(&c.ConfBase)
	}()
	sr.Header, sr.Content, sr.Error = getServiceStatus(&c)
	return
}

func getServiceStatus(c *ConfServices) (header string, content string, err error) {
	t := GetTableWriter(*c.FixedTableWidth)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignRight},
	})

	overall := utils.Good("OK")

	//SERVICES:
	for _, s := range c.Services {
		cmd := exec.Command("/etc/rc.d/rc."+s, "status")
		err := cmd.Run()

		var status string

		if err != nil {
			status = utils.Err("Not running")
			overall = utils.Err("Critical")
		} else {
			status = utils.Good("Running")
		}

		t.AppendRow([]interface{}{s, status})
	}
	
	content = RenderTable(t, "Services: " + overall)
	return
}
