package rules

import (
	"fmt"
	"log/slog"

	"github.com/domsnail/doctryne/pkg/types"
)

type Ruleset struct {
	// Domains
	// todo: must support both .com, .example.com
	Domains []ObjectsRule `yaml:"domains" json:"domains"`

	CountryCodes []ObjectsRule `yaml:"country_codes" json:"country_codes"`
}

type Rule struct {
	ID          string
	Description string

	disabled bool
}

type IRule interface {
	Prepare()
	Validate() error

	IsDisabled() bool
}

type ObjectsRule struct {
	Severity        types.Severity        `yaml:"severity" json:"severity,omitempty"`
	DetectionMethod types.DetectionMethod `yaml:"detection_method" json:"detection_method,omitempty"`

	Allow    []string `yaml:"allow" json:"allow"`
	Disallow []string `yaml:"disallow" json:"disallow"`

	Rule
}

func (r *ObjectsRule) Prepare() {
	if len(r.Allow) == 0 && len(r.Disallow) == 0 {
		slog.Debug(fmt.Sprintf("rule '%s' disabled: no targets", r.ID))
		r.disabled = true
	}

	if r.Severity == "" {
		r.Severity = types.Severity_Low
	}

	if r.DetectionMethod == "" {
		r.DetectionMethod = types.DetectionMethod_Exact
	}
}

func (r *ObjectsRule) Validate() error {
	return nil
}

func (r *ObjectsRule) IsDisabled() bool {
	return r.disabled
}
