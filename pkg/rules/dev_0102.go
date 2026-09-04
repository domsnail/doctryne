package rules

// Rule_DEV_0102 Blacklisted github user ids
type Rule_DEV_0102 struct {
	SlicedRestrictRule `yaml:",inline"`

	UseDatabase bool `json:"use_database" yaml:"use_database"`
}

func (r Rule_DEV_0102) Validate() error {
	return nil
}
