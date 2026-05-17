package cfg

import (
	"errors"
	"net/url"
	"slices"
	"time"

	"github.com/domsnail/doctryne/pkg/types"
)

type Config struct {
	HttpProxy   *url.URL      `json:"http_proxy" yaml:"http_proxy" env:"HTTP_PROXY"`
	Insecure    bool          `json:"insecure" yaml:"insecure" env:"HTTP_INSECURE"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
	CacheMaxAge time.Duration `json:"cache_max_age" yaml:"cache_max_age"`

	Scan   *Scan  `json:"scan" yaml:"scan"`
	Output Output `json:"output" yaml:"output"`

	Credentials Credentials `json:"credentials" yaml:"credentials"`

	Server   Server    `json:"server" yaml:"server" env-prefix:"SRV_"`
	Database *Database `json:"database" yaml:"database" env-prefix:"DB_"`
	Logging  Logging   `json:"logs" yaml:"logs" env-prefix:"LOGS_"`

	FilePath string `json:"-" yaml:"-"`
}

func NewConfigWithDefaultValues() *Config {
	return &Config{
		Insecure:    false,
		Timeout:     time.Second * 30,
		CacheMaxAge: 14 * 24 * time.Hour, // 2 weeks
		Output:      Output{Format: "text"},
		Logging: Logging{
			Format: "text",
		},
		Scan: &Scan{
			ExtractFullContributorInfo: false,
			DeepRepositoryInspection:   false,
		},
	}
}

func (c *Config) HasScan() bool {
	return c.Scan != nil && len(c.Scan.Targets) > 0 && c.Scan.Type != types.ScanType_Unspecified
}

func (c *Config) HasDatabase() bool {
	return c.Database != nil && len(c.Database.Host) > 0 && c.Database.Port > 0
}

func (c *Config) IsValid() error {
	if c.Server.Enabled {
		if c.Server.Host == "" {
			return errors.New("no server host provided")
		}

		if c.Server.Port == 0 {
			return errors.New("no server port provided")
		}

		if !c.HasDatabase() {
			return errors.New("no database connection provided")
		}
	}

	if c.Logging.Format != "text" && c.Logging.Format != "json" {
		return errors.New("invalid logging format, only 'text' and 'json' are supported")
	}

	if !slices.Contains(types.ReportFormats, c.Output.Format) {
		return errors.New("invalid output format")
	}

	return nil
}

type Scan struct {
	Targets []string       `json:"targets" yaml:"targets"`
	Type    types.ScanType `json:"type" yaml:"type"`

	// ExtractFullContributorInfo requests full contributor profile download (from NPM, GitHub or other sources,
	// including other projects, companies etc., may be useful for cross-correlation analysis
	ExtractFullContributorInfo bool `json:"extract_full_contributor_info" yaml:"extract_full_contributor_info"`

	// DeepRepositoryInspection requests full repository source code to analyze source code to find comments and other
	// text structures that might be interesting
	DeepRepositoryInspection bool `json:"deep_repository_inspection" yaml:"deep_repository_inspection"`
}

type Output struct {
	Format types.ReportFormat `json:"format" yaml:"format"`
}

type Credentials struct {
	GithubApiKey string `json:"github_api_key" yaml:"github_api_key" env:"GITHUB_API_KEY"`
	NpmApiKey    string `json:"npm_api_key" yaml:"npm_api_key" env:"NPM_API_KEY"`
}

type Server struct {
	Enabled bool `json:"enabled" yaml:"enabled" env:"ENABLED"`

	Host string `json:"host" yaml:"host" env:"HOST"`
	Port uint32 `json:"port" yaml:"port" env:"PORT"`

	AccessKey string `json:"api_key" yaml:"api_key" env:"ACCESS_KEY"`
}

type Database struct {
	Host     string `json:"host" yaml:"host" env:"HOST"`
	Port     uint32 `json:"port" yaml:"port" env:"PORT"`
	User     string `json:"user" yaml:"user" env:"USER"`
	Name     string `json:"name" yaml:"name" env:"NAME"`
	Timezone string `json:"timezone" yaml:"timezone" env:"TZ"`
}

type Logging struct {
	Level     int    `json:"level" yaml:"level" env:"LEVEL"`
	Format    string `json:"format" yaml:"format" env:"FORMAT"`
	AddSource bool   `json:"add_source" yaml:"add_source" env:"ADD_SOURCE"`
}
