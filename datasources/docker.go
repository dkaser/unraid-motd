package datasources

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/dkaser/unraid-motd/utils"
)

const (
	dockerMinAPI = "1.40"
)

// ConfDocker extends ConfBase with a list of containers to ignore
type ConfDocker struct {
	ConfBase `yaml:",inline"`
	// List of container names to ignore
	Ignore []string `yaml:"ignore,omitempty"`
}

// Init sets up default alignment
func (c *ConfDocker) Init() {
	c.ConfBase.Init()
}

// GetDocker docker container status using the API
func GetDocker(ch chan<- SourceReturn, conf *Conf) {
	c := conf.Docker
	// Check for warnOnly override
	if c.WarnOnly == nil {
		c.WarnOnly = &conf.WarnOnly
	}
	if c.FixedTableWidth == nil {
		c.FixedTableWidth = &conf.FixedTableWidth
	}

	sr := NewSourceReturn(conf.debug)
	defer func() {
		ch <- sr.Return(&c.ConfBase)
	}()
	var err error
	var cl containerList
	cl, err = getDockerContainers()

	if err != nil {
		err = &ModuleNotAvailable{"docker", err}

		t := GetTableWriter(*c.FixedTableWidth)
		sr.Content = RenderTable(t, "Docker: " + utils.Warn("Unavailable"))
	} else {
		sr.Content, sr.Error = cl.getContent(c.Ignore, *c.WarnOnly, *c.FixedTableWidth)
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
