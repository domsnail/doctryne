package cfg

import (
	"time"
)

type VulnerabilityDatabaseConfig struct {
	DisableAutoUpdates    bool `json:"disable_auto_updates" yaml:"disable_auto_updates" env:"DISABLE_AUTO_UPDATES"`
	DisableStartupUpdates bool `json:"disable_startup_updates" yaml:"disable_startup_updates" env:"DISABLE_STARTUP_UPDATES"`

	UpdatesRefresh  time.Duration         `json:"updates_refresh" yaml:"updates_refresh" env-default:"30s"`
	UpdateTimeout   time.Duration         `json:"update_timeout" yaml:"update_timeout" env-default:"8h"`
	UpdateSchedules UpdateSchedulesConfig `json:"update_schedules" yaml:"update_schedules"`

	DisableNVD  bool `json:"disable_nvd" yaml:"disable_nvd" env:"DISABLE_NVD"`
	DisableGhsa bool `json:"disable_ghsa" yaml:"disable_ghsa" env:"DISABLE_GHSA"`
	DisableKev  bool `json:"disable_kev" yaml:"disable_kev" env:"DISABLE_KEV"`
	DisableEpss bool `json:"disable_epss" yaml:"disable_epss" env:"DISABLE_EPSS"`
	DisableOSV  bool `json:"disable_osv" yaml:"disable_osv" env:"DISABLE_OSV"`

	Catalog    string                          `json:"catalog" yaml:"catalog" env-default:"vuln_catalog"`
	GitRemotes VulnerabilityDatabaseGitRemotes `json:"git_remotes" yaml:"git_remotes"`
	Remotes    VulnerabilityDatabaseRemotes    `json:"remotes" yaml:"remotes"`
}

type VulnerabilityDatabaseGitRemotes struct {
	NVD  string `json:"nvd" yaml:"nvd" env-default:"https://github.com/CVEProject/cvelistV5"`
	GHSA string `json:"ghsa" yaml:"ghsa" env-default:"https://github.com/github/advisory-database"`
}

type VulnerabilityDatabaseRemotes struct {
	// todo: use these remotes in download clients
	NVD  string `json:"nvd" yaml:"nvd" env-default:"https://services.nvd.nist.gov/rest/json"`
	GHSA string `json:"ghsa" yaml:"ghsa" env-default:"https://api.github.com/advisories"`

	Epss string `json:"epss" yaml:"epss" env-default:"https://api.first.org/epss"`
}

type UpdateSchedulesConfig struct {
	NVD  string `json:"nvd" yaml:"nvd" env-default:"0 */8 * * *"`
	OSV  string `json:"osv" yaml:"osv" env-default:"0 2 * * *"`
	GHSA string `json:"ghsa" yaml:"ghsa" env-default:"0 */12 * * *"`
	KEV  string `json:"kev" yaml:"kev" env-default:"0 4 * * *"`
	Epss string `json:"epss" yaml:"epss" env-default:"0 6 * * *"`
}
