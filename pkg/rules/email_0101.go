package rules

// Rule_EMAIL_0101 black/whitelisted email addressees
type Rule_EMAIL_0101 struct {
	SlicedRestrictAllowRule `yaml:",inline"`

	UseDatabase bool `json:"use_database" yaml:"use_database"`
}
