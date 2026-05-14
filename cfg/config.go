package cfg

import (
	"errors"
	"net/url"
)

type Config struct {
	HttpProxy *url.URL `json:"http_proxy" yaml:"http_proxy" env:"HTTP_PROXY"`
	Target    string   `json:"target" yaml:"target"`

	Credentials Credentials `json:"credentials" yaml:"credentials"`

	Server  Server  `json:"server" yaml:"server" env-prefix:"SRV_"`
	Logging Logging `json:"logs" yaml:"logs" env-prefix:"LOGS_"`

	FilePath string `json:"-" yaml:"-"`
}

func (c *Config) IsValid() error {
	if c.Server.Enabled {
		if c.Server.Host == "" {
			return errors.New("no server host provided")
		}

		if c.Server.Port == 0 {
			return errors.New("no server port provided")
		}
	} else {
		if c.Target == "" {
			return errors.New("no target provided")
		}
	}

	if c.Logging.Format != "text" && c.Logging.Format != "json" {
		return errors.New("invalid logging format, only 'text' and 'json' are supported")
	}

	return nil
}

type Credentials struct {
	GithubAPIKey string `json:"github_api_key" yaml:"github_api_key" env:"GITHUB_API_KEY"`
}

type Server struct {
	Enabled bool `json:"enabled" yaml:"enabled" env:"ENABLED"`

	Host string `json:"host" yaml:"host" env:"HOST"`
	Port uint32 `json:"port" yaml:"port" env:"PORT"`

	AccessKey string `json:"api_key" yaml:"api_key" env:"ACCESS_KEY"`
}

type Logging struct {
	Level     int    `json:"level" yaml:"level" env:"LEVEL"`
	Format    string `json:"format" yaml:"format" env:"FORMAT"`
	AddSource bool   `json:"add_source" yaml:"add_source" env:"ADD_SOURCE"`
}
