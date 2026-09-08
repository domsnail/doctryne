package cfg

import "time"

type VulnerabilityDatabaseConfig struct {
	UpdatesRefresh  time.Duration         `json:"updates_refresh" yaml:"updates_refresh" env-default:"30s"`
	UpdateTimeout   time.Duration         `json:"update_timeout" yaml:"update_timeout" env-default:"30m"`
	UpdateSchedules UpdateSchedulesConfig `json:"update_schedules" yaml:"update_schedules"`
}

type UpdateSchedulesConfig struct {
	NVD  string `json:"nvd" yaml:"nvd" env-default:"0 */8 * * *"`
	CVE  string `json:"cve" yaml:"cve" env-default:"30 */8 * * *"`
	OSV  string `json:"osv" yaml:"osv" env-default:"0 2 * * *"`
	GHSA string `json:"ghsa" yaml:"ghsa" env-default:"0 */12 * * *"`
	KEV  string `json:"kev" yaml:"kev" env-default:"0 4 * * *"`
}
