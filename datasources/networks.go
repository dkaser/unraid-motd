package datasources

import (
	"github.com/shirou/gopsutil/v3/net"
	"strings"
)

type ConfNet struct {
	ConfBaseWarn `yaml:",inline"`
	IPv4         bool `yaml:"show_ipv4"`
	IPv6         bool `yaml:"show_ipv6"`
}

// Init is mandatory
func (c *ConfNet) Init() {
	// Base init must be called
	c.ConfBaseWarn.Init()

	c.IPv4 = true
	c.IPv6 = false
}

func GetNetworks(channel chan<- SourceReturn, conf *Conf) {
	sourceConf := conf.Networks
	sourceConf.Load(conf)

	returnData := NewSourceReturn(conf.debug)
	defer func() {
		channel <- returnData.Return()
	}()
	returnData.Content, returnData.Error = getNetworkInterfaces(&sourceConf)
}

func getNetworkInterfaces(sourceConf *ConfNet) (content string, err error) {
	outputTable := GetTableWriter(sourceConf)

	deviceIgnore := []string{"lo", "br-", "veth", "docker0", "vnet0"}
	nets, err := net.Interfaces()

INTERFACES:
	for _, net := range nets {
		for _, s := range deviceIgnore {
			if strings.Contains(net.Name, s) {
				continue INTERFACES
			}
		}

		if len(net.Addrs) == 0 {
			continue
		}

		addrs := ""

		for _, addr := range net.Addrs {
			if (strings.Contains(addr.Addr, ".") && sourceConf.IPv4) || (strings.Contains(addr.Addr, ":") && sourceConf.IPv6) {
				addrs += addr.Addr + "\n"
			}
		}

		if addrs != "" {
			outputTable.AppendRow([]interface{}{net.Name, strings.Trim(addrs, "\n")})
		}
	}

	content = RenderTable(outputTable, "Networks")

	return
}
