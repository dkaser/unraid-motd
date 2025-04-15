package datasources

import (
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/dkaser/unraid-motd/utils"
	"gopkg.in/ini.v1"
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

	sshDisabled := true
	smbDisabled := true
	nfsDisabled := true
	dockerDisabled := true

	ident, err := ini.Load("/boot/config/ident.cfg")
	if err == nil {
		sshDisabled = strings.ToLower(ident.Section("").Key("USE_SSH").String()) == "no"
	}

	shares, err := ini.Load("/boot/config/share.cfg")
	if err == nil {
		smbDisabled = strings.ToLower(shares.Section("").Key("shareSMBEnabled").String()) == "no"
		nfsDisabled = strings.ToLower(shares.Section("").Key("shareNFSEnabled").String()) == "no"
	}

	docker, err := ini.Load("/boot/config/docker.cfg")
	if err == nil {
		dockerDisabled = strings.ToLower(docker.Section("").Key("DOCKER_ENABLED").String()) == "no"
	}

	//SERVICES:
	for _, service := range sourceConf.Services {
		reg := regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)
		service = reg.ReplaceAllString(service, "")

		switch service {
		case "sshd":
			if sshDisabled {
				continue
			}
		case "nfsd":
			if nfsDisabled {
				continue
			}
		case "samba":
			if smbDisabled {
				continue
			}
		case "docker":
			if dockerDisabled {
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
