package rules

import "github.com/domsnail/doctryne/pkg/types"

type Ruleset struct {
	// Domains
	// todo: must support both .com, .example.com
	Domains []ObjectsRule `yaml:"domains" json:"domains"`

	CountryCodes []ObjectsRule `yaml:"country_codes" json:"country_codes"`
}

type ObjectsRule struct {
	Severity        types.Severity        `yaml:"severity" json:"severity,omitempty"`
	DetectionMethod types.DetectionMethod `yaml:"detection_method" json:"detection_method,omitempty"`

	Allow    []string `yaml:"allow" json:"allow"`
	Disallow []string `yaml:"disallow" json:"disallow"`
}
