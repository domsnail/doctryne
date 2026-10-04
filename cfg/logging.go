package cfg

type LoggingConfig struct {
	Level     int    `json:"level" yaml:"level" env:"LEVEL" env-default:"0"`
	Format    string `json:"format" yaml:"format" env:"FORMAT" env-default:"text"`
	AddSource bool   `json:"add_source" yaml:"add_source" env:"ADD_SOURCE" env-default:"false"`

	Alerts AlertsConfig `json:"alerts" yaml:"alerts" env:"ALERTS" env-default:""`
}

type AlertsConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled" env:"ENABLED"`

	System AlertsSystem `json:"system" yaml:"system" env:"SYSTEM"`

	Environment string `json:"environment" yaml:"environment" env:"ENVIRONMENT" env-default:"development"`
	Endpoint    string `json:"endpoint" yaml:"endpoint" env:"ENDPOINT"`
	Key         string `json:"key" yaml:"key" env:"KEY"`
	MinLevel    int    `json:"min_level" yaml:"min_level" env:"MIN_LEVEL" env-default:"6"`
}

type AlertsSystem string

const (
	AlertsSystem_GitLab AlertsSystem = "gitlab"
)
