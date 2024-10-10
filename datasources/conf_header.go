package datasources

type ConfHeader struct {
	Show bool `yaml:"show"`
	UseHostname bool `yaml:"use_hostname"`
	CustomText string `yaml:"custom_text"`
	Font string `yaml:"font"`
}

func (c *ConfHeader) Init() {
	c.Show = true
	c.UseHostname = true
	c.CustomText = "Custom"
	c.Font = "Banner.flf"
}