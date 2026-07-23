package cfg

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"slices"
	"time"

	"github.com/domsnail/doctryne/pkg/types"
	"github.com/ilyakaznacheev/cleanenv"
)

var GlobalConfig *Config

type Config struct {
	HttpProxy    string        `json:"http_proxy" yaml:"http_proxy" env:"HTTP_PROXY"`
	Insecure     bool          `json:"insecure" yaml:"insecure" env:"HTTP_INSECURE"`
	Timeout      time.Duration `json:"timeout" yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"30s"`
	AllowedHosts []string      `json:"allowed_hosts" yaml:"allowed_hosts" env:"HTTP_ALLOWED_HOSTS" env-separator:","`

	CacheMaxAge time.Duration `json:"cache_max_age" yaml:"cache_max_age" env:"CACHE_MAX_AGE" env-default:"5m"`
	Concurrency int32         `json:"concurrency" yaml:"concurrency" env:"CONCURRENCY" env-default:"4"`

	// If ServerURL present, cli will use remote server rpc to create new inspection
	ServerURL string `json:"server_url" yaml:"server_url"`

	Scan       ScanConfig                 `json:"scan" yaml:"scan"`
	Languages  LanguagesConfig            `json:"languages" yaml:"languages"`
	GitHistory GitHistoryInspectionConfig `json:"git_history" yaml:"git_history"`

	Output Output `json:"output" yaml:"output"`

	Credentials Credentials `json:"credentials" yaml:"credentials"`
	RateLimits  RateLimits  `json:"rate_limits" yaml:"rate_limits"`

	Server   Server    `json:"server" yaml:"server" env-prefix:"SRV_"`
	Database *Database `json:"database" yaml:"database" env-prefix:"DB_"`
	Logging  Logging   `json:"logs" yaml:"logs" env-prefix:"LOGS_"`

	FilePath string `json:"-" yaml:"-"`
}

func SetGlobalConfig(cfg *Config) {
	GlobalConfig = cfg

	setGlobalHttpClient(cfg)
}

func setGlobalHttpClient(cfg *Config) {
	if cfg.HttpProxy != "" {
		proxyURL, err := url.Parse(cfg.HttpProxy)
		if err != nil {
			panic(err)
		}

		http.DefaultTransport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.Insecure,
			},
		}

		http.DefaultClient.Transport = http.DefaultTransport
	}

	http.DefaultClient.Timeout = cfg.Timeout
}

func NewConfigWithDefaultValues() *Config {
	return &Config{
		Insecure:    false,
		Timeout:     time.Second * 30,
		CacheMaxAge: 14 * 24 * time.Hour, // 2 weeks
		Output:      Output{Format: types.ReportFormat_TextTable},
		Concurrency: int32(runtime.NumCPU()),
		Logging: Logging{
			Format: "text",
		},
		Languages: LanguagesConfig{
			JavaScript: JavaScriptConfig{
				CheckOptionalDependencies: false,
				CheckDevDependencies:      false,
			},
		},
		Scan: ScanConfig{
			ExtractFullContributorInfo: false,
			DeepRepositoryInspection:   false,
			FileSearchDepth:            10,
		},
	}
}

func NewConfigFromFile(filepath string) (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig(filepath, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func NewConfigFromEnv() (*Config, error) {
	var cfg Config

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) HasScan() bool {
	return len(c.Scan.Targets) > 0 && c.Scan.Type != types.ScanType_Unspecified
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

		//if !c.HasDatabase() { todo
		//	return errors.New("no database connection provided")
		//}
	} else if c.ServerURL != "" {
		_, err := url.Parse(c.ServerURL)
		return fmt.Errorf("invalid server url: %w", err)
	}

	if c.Logging.Format != "text" && c.Logging.Format != "json" {
		return errors.New("invalid logging format, only 'text' and 'json' are supported")
	}

	if c.Concurrency <= 0 {
		return errors.New("invalid concurrency modifier: cannot be less than zero")
	}

	if !slices.Contains(types.ReportFormats, c.Output.Format) {
		return errors.New("invalid output format")
	}

	return nil
}

// ScanConfig defines default scanning properties for all setups,
// will be overridden by user provided values
type ScanConfig struct {
	Targets []string       `json:"targets" yaml:"targets"`
	Type    types.ScanType `json:"type" yaml:"type"`

	// ExtractFullContributorInfo requests full contributor profile download (from NPM, GitHub or other sources,
	// including other projects, companies etc., may be useful for cross-correlation analysis
	ExtractFullContributorInfo bool `json:"extract_full_contributor_info" yaml:"extract_full_contributor_info"`

	// DeepRepositoryInspection requests full repository source code to analyze source code to find comments and other
	// text structures that might be interesting
	DeepRepositoryInspection bool `json:"deep_repository_inspection" yaml:"deep_repository_inspection"`

	// FileSearchDepth sets maximum file directory depth when analyzing directory,
	// cannot be overridden
	FileSearchDepth int `json:"file_search_depth" yaml:"file_search_depth"`

	// ActivityPeriod defines depth of analysis of latest activity on scanned objects (repositories, registry stats, etc.),
	// default is 3 month (2160 hours)
	ActivityPeriod time.Duration `json:"activity_period" yaml:"activity_period" env-default:"2160h"`
}

type Output struct {
	Format types.ReportFormat `json:"format" yaml:"format"`
}

type Credentials struct {
	GithubApiKey        string `json:"github_api_key" yaml:"github_api_key" env:"GITHUB_API_KEY"`
	NpmApiKey           string `json:"npm_api_key" yaml:"npm_api_key" env:"NPM_API_KEY"`
	StackExchangeApiKey string `json:"stack_exchange_api_key" yaml:"stack_exchange_api_key" env:"STACK_EXCHANGE_API_KEY"`
}

type Server struct {
	Enabled bool `json:"enabled" yaml:"enabled" env:"ENABLED"`

	Host string `json:"host" yaml:"host" env:"HOST" env-default:"0.0.0.0"`
	Port uint32 `json:"port" yaml:"port" env:"PORT" env-default:"8080"`

	DisableHealth  bool `json:"disable_health" yaml:"disable_health"`
	DisableMetrics bool `json:"disable_metrics" yaml:"disable_metrics"`
	DisableReflect bool `json:"disable_reflect" yaml:"disable_reflect"`

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

type RateLimits struct {
	RefreshPeriod time.Duration `json:"refresh_period" yaml:"refresh_period" env-default:"1m"`

	// Amount of requests that can be sent to GitHub API in 1m when authorized with PAT,
	// refer to: https://docs.github.com/ru/rest/using-the-rest-api/rate-limits-for-the-rest-api
	GitHub uint64 `json:"github" yaml:"github" env-default:"900"`

	NPM uint64 `json:"npm" yaml:"npm" env-default:"1000"`
}
