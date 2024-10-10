package datasources

import (
	"github.com/dkaser/unraid-motd/utils"
	"os/exec"
)

type ConfServices struct {
	ConfBase `yaml:",inline"`
	Services []string `yaml:"monitor,flow"`
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
	c.Load(conf)

	sr := NewSourceReturn(conf.debug)
	defer func() {
		ch <- sr.Return(&c.ConfBase)
	}()
	sr.Content, sr.Error = getServiceStatus(&c)
	return
}

func getServiceStatus(c *ConfServices) (content string, err error) {
	t := GetTableWriter(c)

	overall := utils.Good("OK")

	//SERVICES:
	for _, s := range c.Services {
		cmd := exec.Command("/etc/rc.d/rc."+s, "status")
		err := cmd.Run()

		if err != nil {
			overall = utils.Err("Critical")
			t.AppendRow([]interface{}{s, utils.Err("Not running")})
		} else if (! *c.WarnOnly) {
			t.AppendRow([]interface{}{s, utils.Good("Running")})
		}
	}

	content = RenderTable(t, "Services: " + overall)
	return
}
