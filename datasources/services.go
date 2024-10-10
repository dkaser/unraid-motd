package datasources

import (
	"github.com/dkaser/unraid-motd/utils"
	"os/exec"
	"regexp"
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

func GetServices(channel chan<- SourceReturn, conf *Conf) {
	sourceConf := conf.Services
	sourceConf.Load(conf)

	returnData := NewSourceReturn(conf.debug)
	defer func() {
		channel <- returnData.Return()
	}()
	returnData.Content = getServiceStatus(&sourceConf)
}

func getServiceStatus(sourceConf *ConfServices) (content string) {
	outputTable := GetTableWriter(sourceConf)

	overall := utils.Good("OK")

	//SERVICES:
	for _, service := range sourceConf.Services {
		reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)
		service = reg.ReplaceAllString(service, "")

		cmd := exec.Command("/etc/rc.d/rc."+service, "status") // #nosec G204
		err := cmd.Run()

		if err != nil {
			overall = utils.Err("Critical")
			outputTable.AppendRow([]interface{}{service, utils.Err("Not running")})
		} else if !*sourceConf.WarnOnly {
			outputTable.AppendRow([]interface{}{service, utils.Good("Running")})
		}
	}

	content = RenderTable(outputTable, "Services: "+overall)

	return
}
