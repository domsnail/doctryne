package rules

// Rule_PKG_0102 black/whitelisted package vendors
type Rule_PKG_0102 struct {
	SlicedRestrictRule `yaml:",inline"`

	UseDatabase bool `json:"use_database" yaml:"use_database"`
}

func (r Rule_PKG_0102) Validate() error {
	return nil
}
