package cfg

import (
	"errors"
	"net/url"
)

type Config struct {
	HttpProxy *url.URL `json:"http_proxy" yaml:"http_proxy" env:"HTTP_PROXY"`
	Target    string   `json:"target" yaml:"target"`

	Credentials Credentials `json:"credentials" yaml:"credentials"`

	Server Server `json:"server" yaml:"server" env-prefix:"SERVER_"`
}

func (c *Config) Defaults() *Config {
	return c
}

func (c *Config) IsValid() (bool, error) {
	if c.Target == "" {
		return false, errors.New("no target provided")
	}

	if c.Server.Enabled {
		if c.Server.Host == "" {
			return false, errors.New("no server host provided")
		}

		if c.Server.Port == 0 {
			return false, errors.New("no server port provided")
		}
	}

	return true, nil
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
