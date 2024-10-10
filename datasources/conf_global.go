package datasources

// ConfGlobal is the config struct for global settings
type ConfGlobal struct {
	// Hide fields which are deemed to be OK
	WarnOnly bool `yaml:"warnings_only"`
	// Define how data sources are displayed
	ColDef [][]string `yaml:"display,flow,omitempty"`
	// Padding between columns when using col_def
	ColPad int `yaml:"padding"`
	// Internal variables
	debug bool

	FixedTableWidth int `yaml:"table_width"`
	Border bool `yaml:"border"`
}
