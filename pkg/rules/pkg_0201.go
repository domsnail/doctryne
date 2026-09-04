package rules

// Rule_PKG_0201 requires time to pass after package version release
type Rule_PKG_0201 struct {
	TimePassedRule `yaml:",inline"`
}
