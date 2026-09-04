package rules

// Rule_DEV_0202 black/whitelisted developer organizations
type Rule_DEV_0202 struct {
	SlicedRestrictAllowRule `yaml:",inline"`

	UseDatabase bool `json:"use_database" yaml:"use_database"`
}
