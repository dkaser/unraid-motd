package datasources

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/cosandr/go-motd/utils"
)

const (
	dockerMinAPI = "1.40"
)

// ConfDocker extends ConfBase with a list of containers to ignore
type ConfDocker struct {
	ConfBase `yaml:",inline"`
	// Interact directly with the docker CLI, much slower than API
	Exec bool `yaml:"use_exec"`
	// List of container names to ignore
	Ignore []string `yaml:"ignore,omitempty"`
}

// Init sets up default alignment
func (c *ConfDocker) Init() {
	c.ConfBase.Init()
	c.PadHeader[1] = 4
	c.PadContent[1] = 1
}

// GetDocker docker container status using the API
func GetDocker(ch chan<- SourceReturn, conf *Conf) {
	c := conf.Docker
	// Check for warnOnly override
	if c.WarnOnly == nil {
		c.WarnOnly = &conf.WarnOnly
	}
	sr := NewSourceReturn(conf.debug)
	defer func() {
		ch <- sr.Return(&c.ConfBase)
	}()
	var err error
	var cl containerList
	if c.Exec {
		cl, err = getContainersExec(false, false)
	} else {
		cl, err = getDockerContainers()
	}
	if err != nil {
		err = &ModuleNotAvailable{"docker", err}
		sr.Header = fmt.Sprintf("%s: %s\n", utils.Wrap("Docker", c.padL, c.padR), utils.Warn("unavailable"))
	} else {
		sr.Header, sr.Content, sr.Error = cl.toHeaderContent(c.Ignore, *c.WarnOnly, c.padL, c.padR)
	}
}

func getDockerContainers() (cl containerList, err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion(dockerMinAPI))
	if err != nil {
		return
	}

	allContainers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return
	}
	cl.Runtime = "Docker"
	cl.Root = true
	for _, container := range allContainers {
		cl.Containers = append(cl.Containers, containerStatus{
			Name:   strings.TrimPrefix(container.Names[0], "/"),
			Status: container.State,
		})
	}
	return
}
