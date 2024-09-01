package conf

type DatabaseConf struct {
	Driver            string `json:"driver" mapstructure:"driver" yaml:"driver"`
	Connection        string `json:"connection" mapstructure:"connection" yaml:"connection"`
	MaxIdleConnection int    `json:"max_idle" mapstructure:"max_idle" yaml:"max_idle"`
	MaxOpenConnection int    `json:"max_open" mapstructure:"max_open" yaml:"max_open"`
}
