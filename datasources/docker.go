package datasources

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
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
	Ignore []string `yaml:"ignore"`
}

// Init sets up default alignment
func (c *ConfDocker) Init() {
	c.ConfBase.Init()
	c.Ignore = []string{}
}

// GetDocker docker container status using the API
func GetDocker(channel chan<- SourceReturn, conf *Conf) {
	sourceConf := conf.Docker
	sourceConf.Load(conf)

	returnData := NewSourceReturn(conf.debug)
	defer func() {
		channel <- returnData.Return()
	}()
	var err error
	var containers containerList
	containers, err = getDockerContainers()

	if err != nil {
		t := GetTableWriter(sourceConf)
		returnData.Content = RenderTable(t, "Docker: "+utils.Warn("Unavailable"))
	} else {
		returnData.Content = containers.getContent(sourceConf.Ignore, *sourceConf.WarnOnly, sourceConf)
	}
}

func getDockerContainers() (containers containerList, err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion(dockerMinAPI))
	if err != nil {
		return
	}

	allContainers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return
	}
	containers.Runtime = "Docker"
	containers.Root = true
	for _, container := range allContainers {
		containers.Containers = append(containers.Containers, containerStatus{
			Name:   dockerContainerName(container),
			Status: container.State,
		})
	}

	return
}

func dockerContainerName(container types.Container) string {
	for _, name := range container.Names {
		name = strings.TrimPrefix(strings.TrimSpace(name), "/")
		if name != "" {
			return name
		}
	}

	id := strings.TrimSpace(container.ID)
	if len(id) > 12 {
		id = id[:12]
	}
	if id != "" {
		return id
	}

	return "unknown"
}
