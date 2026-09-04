package rules

// Rule_PKG_0101 black/whitelisted package names
type Rule_PKG_0101 struct {
	SlicedRestrictRule `yaml:",inline"`

	UseDatabase bool `json:"use_database" yaml:"use_database"`
}

func (r Rule_PKG_0101) Validate() error {
	return nil
}
