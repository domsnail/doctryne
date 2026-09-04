package rules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

type Ruleset struct {
	Developers *DevelopersRuleset `json:"developers,omitempty" yaml:"developers,omitempty"`
	Emails     *EmailsRuleset     `json:"email,omitempty" yaml:"email,omitempty"`
	Packages   *PackagesRuleset   `json:"packages,omitempty" yaml:"packages,omitempty"`
}

func NewRulesetFromFile(path string) (*Ruleset, error) {
	file, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var set Ruleset

	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		decoder := yaml.NewDecoder(bytes.NewReader(file))
		decoder.KnownFields(true)

		err = decoder.Decode(&set)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal yaml file: %w", err)
		}
	case ".json":
		decoder := json.NewDecoder(bytes.NewBuffer(file))
		decoder.DisallowUnknownFields()

		err = decoder.Decode(&set)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal json file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", filepath.Ext(path))
	}

	return &set, nil
}

type DevelopersRuleset struct {
	Usernames *Rule_DEV_0101 `json:"dev_0101,omitempty" yaml:"dev_0101,omitempty"`

	// Github
	GithubID *Rule_DEV_0102 `json:"dev_0102,omitempty" yaml:"dev_0102,omitempty"`

	Location     *Rule_DEV_0201 `json:"dev_0201,omitempty" yaml:"dev_0201,omitempty"`
	Organization *Rule_DEV_0202 `json:"dev_0202,omitempty" yaml:"dev_0202,omitempty"`
}

type EmailsRuleset struct {
	Email *Rule_EMAIL_0101 `json:"email_0101,omitempty" yaml:"email_0101,omitempty"`
}

type PackagesRuleset struct {
	Names   *Rule_PKG_0101 `json:"pkg_0101,omitempty" yaml:"pkg_0101,omitempty"`
	Vendors *Rule_PKG_0102 `json:"pkg_0102,omitempty" yaml:"pkg_0102,omitempty"`

	VersionTimePassed *Rule_PKG_0201 `json:"pkg_0201,omitempty" yaml:"pkg_0201,omitempty"`
}

type CVEsRuleset struct {
	MaxCVSS   *Rule_CVE_0101 `json:"cve_0101,omitempty" yaml:"cve_0101,omitempty"`
	TotalCVEs *Rule_CVE_0102 `json:"cve_0102,omitempty" yaml:"cve_0102,omitempty"`
	TotalCVSS *Rule_CVE_0103 `json:"cve_0103,omitempty" yaml:"cve_0103,omitempty"`
}
