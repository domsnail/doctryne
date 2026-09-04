package rules

// Rule_CVE_0103 black/whitelisted CWE list
type Rule_CVE_0104 struct {
	SlicedRestrictAllowRule `yaml:",inline"`

	UseDatabase bool `json:"use_database" yaml:"use_database"`
}
