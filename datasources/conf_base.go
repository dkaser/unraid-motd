package datasources

// ConfBase is the common type for all modules
//
// Custom modules should respect these options
type ConfBase struct {
	// Override global setting
	WarnOnly        *bool `yaml:"warnings_only,omitempty"`
	FixedTableWidth *int  `yaml:"table_width,omitempty"`
	Border          *bool `yaml:"border,omitempty"`
}

func (c *ConfBase) Init() {
}

func (c ConfBase) GetBorder() bool {
	return *c.Border
}

func (c ConfBase) GetTableWidth() int {
	return *c.FixedTableWidth
}

func (c *ConfBase) Load(conf *Conf) {
	if c.FixedTableWidth == nil {
		c.FixedTableWidth = &conf.FixedTableWidth
	}
	if c.WarnOnly == nil {
		c.WarnOnly = &conf.WarnOnly
	}
	if c.Border == nil {
		c.Border = &conf.Border
	}
}
