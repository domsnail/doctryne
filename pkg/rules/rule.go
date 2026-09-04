package rules

import (
	"time"
)

type IRule interface {
	Validate() error
	Prepare() error

	IsDisabled() bool
}

type Rule struct {
	ID          string  `json:"id" yaml:"id"`
	Description string  `json:"description" yaml:"description"`
	CVSS        float32 `json:"cvss,omitempty" yaml:"cvss,omitempty"`
	CWE         string  `json:"cwe,omitempty" yaml:"cwe,omitempty"`

	Disabled bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`
}

type DurationRule struct {
	Rule

	Duration time.Duration `json:"duration" yaml:"duration"`
}

type NumericRule struct {
	Rule

	Max uint64 `json:"max" yaml:"max"`
	Min uint64 `json:"min" yaml:"min"`
}

func (r Rule) IsDisabled() bool {
	return r.Disabled
}
