package cfg

import (
	"net/url"
	"time"
)

type VulnerabilityDatabaseConfig struct {
	DisableAutoUpdates bool `json:"disable_auto_updates" yaml:"disable_auto_updates"`

	UpdatesRefresh  time.Duration         `json:"updates_refresh" yaml:"updates_refresh" env-default:"30s"`
	UpdateTimeout   time.Duration         `json:"update_timeout" yaml:"update_timeout" env-default:"30m"`
	UpdateSchedules UpdateSchedulesConfig `json:"update_schedules" yaml:"update_schedules"`

	Catalog         string                               `json:"catalog" yaml:"catalog" env-default:"vuln_catalog"`
	GitRemotes      VulnerabilityDatabaseGitRemotes      `json:"git_remotes" yaml:"git_remotes"`
	DownloadRemotes VulnerabilityDatabaseDownloadRemotes `json:"download_remotes" yaml:"download_remotes"`
}

type VulnerabilityDatabaseGitRemotes struct {
	NVD  url.URL `json:"nvd" yaml:"nvd" env-default:"https://github.com/CVEProject/cvelistV5"`
	GHSA url.URL `json:"ghsa" yaml:"ghsa" env-default:"https://github.com/github/advisory-database"`
}

type VulnerabilityDatabaseDownloadRemotes struct {
	NVD  url.URL `json:"nvd" yaml:"nvd" env-default:"https//github.com/CVEProject/cvelistV5/archive/refs/heads/main.zip"`
	GHSA url.URL `json:"ghsa" yaml:"ghsa" env-default:"https://github.com/github/advisory-database/archive/refs/heads/main.zip"`
}

type UpdateSchedulesConfig struct {
	NVD  string `json:"nvd" yaml:"nvd" env-default:"0 */8 * * *"`
	CVE  string `json:"cve" yaml:"cve" env-default:"30 */8 * * *"`
	OSV  string `json:"osv" yaml:"osv" env-default:"0 2 * * *"`
	GHSA string `json:"ghsa" yaml:"ghsa" env-default:"0 */12 * * *"`
	KEV  string `json:"kev" yaml:"kev" env-default:"0 4 * * *"`
}
