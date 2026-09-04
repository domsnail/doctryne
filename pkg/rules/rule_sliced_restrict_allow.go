package rules

import (
	"fmt"
	"regexp"
	"strings"
)

type SlicedRestrictAllowRule struct {
	SlicedRestrictRule `yaml:",inline"`

	Allow []string `json:"allow,omitempty" yaml:"allow,omitempty"`
}

func (r *SlicedRestrictAllowRule) Validate() error {
	return nil
}

func (r *SlicedRestrictAllowRule) Prepare() error {
	var (
		allowRegexps, restrictRegexps   []*regexp.Regexp
		allowContains, restrictContains []string
		allowExacts, restrictExacts     []string

		err error
	)

	if len(r.Allow) == 0 && len(r.Restrict) == 0 {
		r.Exec = func(value string) (passed bool, violation string) {
			return true, ""
		}

		return nil
	}

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

	if len(r.Allow) == 0 {
		return r.buildBlacklistFunc(restrictRegexps, restrictContains, restrictExacts)
	}

	if len(r.Allow) > 0 {
		allowRegexps, allowContains, allowExacts, err = r.prepareValues(r.Allow)
		if err != nil {
			return fmt.Errorf("failed to prepare allowed values: %w", err)
		}
	}

	r.Exec = func(value string) (passed bool, violation string) {
		for _, v := range allowExacts {
			if strings.EqualFold(value, v) {
				return true, ""
			}
		}

		for _, v := range allowContains {
			if strings.Contains(value, v) {
				return true, ""
			}
		}

		for _, rx := range allowRegexps {
			if rx.MatchString(value) {
				return true, ""
			}
		}

		for _, v := range restrictExacts {
			if strings.EqualFold(value, v) {
				return false, "matched exact blacklisted statement"
			}
		}

		for _, v := range restrictContains {
			if strings.Contains(value, v) {
				return false, "matched blacklisted contains statement"
			}
		}

		for _, rx := range restrictRegexps {
			if rx.MatchString(value) {
				return false, "matched blacklisted regexp"
			}
		}

		return true, ""
	}

	return nil
}

func (r *SlicedRestrictAllowRule) whitelistFunc(regexps []*regexp.Regexp, contains []string, exacts []string) error {
	r.Exec = func(value string) (passed bool, violation string) {
		for _, v := range exacts {
			if strings.EqualFold(value, v) {
				return true, ""
			}
		}

		for _, v := range contains {
			if strings.Contains(value, v) {
				return true, ""
			}
		}

		for _, rx := range regexps {
			if rx.MatchString(value) {
				return true, ""
			}
		}

		return false, ""
	}

	return nil
}
