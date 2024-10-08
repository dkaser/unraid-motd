package datasources

import (
  "github.com/shirou/gopsutil/v3/net"
  "github.com/jedib0t/go-pretty/v6/table"
  "github.com/jedib0t/go-pretty/v6/text"
  "strings"
)

type ConfNet struct {
	ConfBaseWarn `yaml:",inline"`
}

// Init is mandatory
func (c *ConfNet) Init() {
    // Base init must be called
    c.ConfBaseWarn.Init()
    c.PadHeader[1] = 1
	c.PadContent[1] = 1
	c.WarnOnly = new(bool)
}

func GetNetworks(ch chan<- SourceReturn, conf *Conf) {
	c := conf.Networks
	sr := NewSourceReturn(conf.debug)
	defer func() {
		ch <- sr.Return(&c.ConfBase)
	}()
    sr.Header, sr.Content, sr.Error = getNetworkInterfaces(&c)
    return
}

func getNetworkInterfaces(c *ConfNet) (header string, content string, err error) {
	deviceIgnore := []string{"lo", "br-", "veth", "docker0", "vnet0" }

	nets, err := net.Interfaces()
	t := table.NewWriter()

	INTERFACES:
	for _, n := range nets {

		for _, s := range deviceIgnore {
			if(strings.Contains(n.Name, s)) {
				continue INTERFACES
			}
		}

		if len(n.Addrs) == 0 {
			continue
		}

		addrs := ""

		for _, a := range n.Addrs {
			addrs += a.Addr + "\n"
		}

		t.AppendRow([]interface{}{n.Name, strings.Trim(addrs, "\n")})
	}

	t.SetStyle(GetTableStyle())
	t.SetColumnConfigs([]table.ColumnConfig{
        {Number: 1, Align: text.AlignRight},
    })
	t.SetTitle("Networks")

	content = t.Render();
	return
}