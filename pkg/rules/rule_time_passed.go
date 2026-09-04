package rules

import (
	"encoding/json"
	"fmt"
	"time"
)

type TimePassedRule struct {
	Rule `yaml:",inline"`

	// return error if Period not yet passed
	Period time.Duration `json:"period" yaml:"period"`

	Exec func(value time.Time) (passed bool, violation string) `json:"-" yaml:"-"`
}

func (r *TimePassedRule) Validate() error {
	if r.Disabled {
		return nil
	}

	if r.Period <= 0 {
		return fmt.Errorf("period cannot be less or equal to zero")
	}

	return nil
}

func (r *TimePassedRule) Prepare() error {
	r.Exec = func(value time.Time) (bool, string) {
		if time.Now().Add(-r.Period).After(value) {
			return true, ""
		}

		limit := time.Now().Add(-r.Period)
		return false, fmt.Sprintf("required amount of time has not been passed: time left %s", value.Sub(limit).String())
	}

	return nil
}

func (r *TimePassedRule) UnmarshalJSON(data []byte) error {
	type Alias TimePassedRule

	aux := struct {
		Timeout string `json:"period"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	duration, err := time.ParseDuration(aux.Timeout)
	if err != nil {
		return fmt.Errorf("parse timeout: %w", err)
	}

	r.Period = duration
	return nil
}
