package rules

// Rule_DEV_0101 black/whitelisted developer usernames
type Rule_DEV_0101 struct {
	SlicedRestrictRule `yaml:",inline"`

	UseDatabase bool `json:"use_database" yaml:"use_database"`
}

func (r Rule_DEV_0101) Validate() error {
	return nil
}
