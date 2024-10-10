package datasources

// Conf is the combined config struct, defines YAML file
type Conf struct {
	ConfGlobal   `yaml:"global"`
	Header       ConfHeader   `yaml:"header"`
	CPU          ConfTempCPU  `yaml:"cpu"`
	Docker       ConfDocker   `yaml:"docker"`
	SysInfo      ConfSysInfo  `yaml:"sysinfo"`
	UserDrives   ConfDrives   `yaml:"user-drives"`
	SystemDrives ConfDrives   `yaml:"system-drives"`
	Networks     ConfNet      `yaml:"network"`
	Services     ConfServices `yaml:"services"`
}

// Init a config with sane default values
func (c *Conf) Init() {
	// Set global defaults
	c.WarnOnly = true
	c.Border = true
	c.ColPad = 1
	c.ColDef = [][]string{
		{"sysinfo"},
		{"docker", "cpu"},
		{"services", "networks"},
		{"user-drives", "system-drives"},
	}
	c.FixedTableWidth = 60
	c.Header.Init()
	// Init data source configs
	c.CPU.Init()
	c.Docker.Init()
	c.SysInfo.Init()
	c.UserDrives.Init()
	c.SystemDrives.Init()
	c.Networks.Init()
	c.Services.Init()
}
