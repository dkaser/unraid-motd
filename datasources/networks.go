package datasources

import (
	"github.com/shirou/gopsutil/v3/net"
	"strings"
)

type ConfNet struct {
	ConfBaseWarn `yaml:",inline"`
	IPv4 bool `yaml:"show_ipv4"`
	IPv6 bool `yaml:"show_ipv6"`
}

// Init is mandatory
func (c *ConfNet) Init() {
	// Base init must be called
	c.ConfBaseWarn.Init()
	
	c.IPv4 = true
	c.IPv6 = false

}

func GetNetworks(ch chan<- SourceReturn, conf *Conf) {
	c := conf.Networks
	if c.FixedTableWidth == nil {
		c.FixedTableWidth = &conf.FixedTableWidth
	}

	sr := NewSourceReturn(conf.debug)
	defer func() {
		ch <- sr.Return(&c.ConfBase)
	}()
	sr.Content, sr.Error = getNetworkInterfaces(&c)
	return
}

func getNetworkInterfaces(c *ConfNet) (content string, err error) {
	t := GetTableWriter(*c.FixedTableWidth)

	deviceIgnore := []string{"lo", "br-", "veth", "docker0", "vnet0"}
	nets, err := net.Interfaces()

INTERFACES:
	for _, n := range nets {

		for _, s := range deviceIgnore {
			if strings.Contains(n.Name, s) {
				continue INTERFACES
			}
		}

		if len(n.Addrs) == 0 {
			continue
		}

		addrs := ""

		for _, a := range n.Addrs {
			if (strings.Contains(a.Addr, ".") && c.IPv4) || (strings.Contains(a.Addr, ":") && c.IPv6) {
				addrs += a.Addr + "\n"
			}
		}

		if (addrs != "") {
			t.AppendRow([]interface{}{n.Name, strings.Trim(addrs, "\n") })
		}
	}

	content = RenderTable(t, "Networks")
	return
}
