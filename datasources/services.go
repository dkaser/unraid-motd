package datasources

import (
	"github.com/dkaser/unraid-motd/utils"
	"os/exec"
	"regexp"
	"os"
	"errors"
	"gopkg.in/ini.v1"
	"strings"
	"fmt"
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

	ident, err := ini.Load("/boot/config/ident.cfg")
    if err != nil {
        fmt.Printf("Fail to read file: %v", err)
    }

	shares, err := ini.Load("/boot/config/share.cfg")
    if err != nil {
        fmt.Printf("Fail to read file: %v", err)
    }

	docker, err := ini.Load("/boot/config/docker.cfg")
    if err != nil {
        fmt.Printf("Fail to read file: %v", err)
    }

	//SERVICES:
	for _, service := range sourceConf.Services {
		reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)
		service = reg.ReplaceAllString(service, "")

		switch service {
			case "sshd":
				if (strings.ToLower(ident.Section("").Key("USE_SSH").String()) == "no") {
					continue
				}
			case "nfsd":
				if (strings.ToLower(shares.Section("").Key("shareNFSEnabled").String()) == "no") {
					continue
				}
			case "samba":
				if (strings.ToLower(shares.Section("").Key("shareSMBEnabled").String()) == "no") {
					continue
				}
			case "docker":
				if (strings.ToLower(docker.Section("").Key("DOCKER_ENABLED").String()) == "no") {
					continue
				}
		}

		cmd := exec.Command("/etc/rc.d/rc."+service, "status") // #nosec G204
		err := cmd.Run()

		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			overall = utils.Err("Critical")
			outputTable.AppendRow([]interface{}{service, utils.Err("Not running")})
		} else if !*sourceConf.WarnOnly {
			outputTable.AppendRow([]interface{}{service, utils.Good("Running")})
		}
	}

	content = RenderTable(outputTable, "Services: "+overall)

	return
}
