package rules

import (
	"fmt"
	"regexp"
	"strings"
)

type SlicedRestrictRule struct {
	Rule `yaml:",inline"`

	Restrict []string `json:"restrict,omitempty" yaml:"restrict,omitempty"`

	Exec func(value string) (passed bool, violation string) `json:"-" yaml:"-"`
}

func (r *SlicedRestrictRule) Prepare() error {
	var (
		restrictRegexps  []*regexp.Regexp
		restrictContains []string
		restrictExacts   []string

		err error
	)

	if len(r.Restrict) > 0 {
		restrictRegexps, restrictContains, restrictExacts, err = r.prepareValues(r.Restrict)
		if err != nil {
			return fmt.Errorf("failed to prepare restricted values: %w", err)
		}
	} else {
		r.Exec = func(value string) (passed bool, violation string) {
			return true, ""
		}

		return nil
	}

	return r.buildBlacklistFunc(restrictRegexps, restrictContains, restrictExacts)
}

func (r *SlicedRestrictRule) prepareValues(values []string) (regexps []*regexp.Regexp, contains []string, exacts []string, err error) {
	for i, v := range values {
		parts := strings.Split(v, ":")
		switch parts[0] {
		case "regexp":
			regx, err := regexp.Compile(strings.TrimPrefix(v, "regexp:"))
			if err != nil {
				return nil, nil, nil, fmt.Errorf("invalid regexp on row %d: %s", i, v)
			}

			regexps = append(regexps, regx)
		case "contains":
			contains = append(contains, strings.TrimPrefix(v, "contains:"))
		case "exact":
			exacts = append(exacts, strings.TrimPrefix(v, "exact:"))
		default:
			exacts = append(exacts, v)
		}
	}

	return regexps, contains, exacts, nil
}

func (r *SlicedRestrictRule) buildBlacklistFunc(regexps []*regexp.Regexp, contains []string, exacts []string) error {
	r.Exec = func(value string) (passed bool, violation string) {
		for _, rx := range regexps {
			if rx.MatchString(value) {
				return false, "matched blacklisted regexp"
			}
		}

		for _, v := range contains {
			if strings.Contains(value, v) {
				return false, "matched blacklisted contains statement"
			}
		}

		for _, v := range exacts {
			if strings.EqualFold(value, v) {
				return false, "matched exact blacklisted statement"
			}
		}

		return true, ""
	}

	return nil
}
