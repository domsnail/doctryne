package rules

// Rule_DEV_0201 blacklisted developer locations
type Rule_DEV_0201 struct {
	SlicedRestrictRule `yaml:",inline"`

	UseDatabase bool `json:"use_database" yaml:"use_database"`
}

func (r Rule_DEV_0201) Validate() error {
	return nil
}
