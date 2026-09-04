package rules

import (
	"errors"
)

type NumericRangeRule struct {
	Rule `yaml:",inline"`

	Min *float32 `json:"min,omitempty" yaml:"min,omitempty"`
	Max *float32 `json:"max,omitempty" yaml:"max,omitempty"`

	Exec func(value float32) (passed bool, violation string) `json:"-" yaml:"-"`
}

func (r *NumericRangeRule) Validate() error {
	if r.Disabled {
		return nil
	}

	if r.Min == nil && r.Max == nil {
		return errors.New("min and/or max values must be set")
	}

	if r.Min != nil && r.Max != nil {
		if *r.Min == *r.Max {
			return errors.New("min value can not be equal to max value")
		}

		if *r.Min > *r.Max {
			return errors.New("max value must be greater than min value")
		}
	}

	return nil
}

func (r *NumericRangeRule) Prepare() error {
	if r.Min != nil && r.Max != nil {
		r.Exec = func(value float32) (passed bool, violation string) {
			if *r.Min < value && value < *r.Max {
				return true, ""
			}

			return false, "value out of range"
		}

		return nil
	}

	if r.Min != nil {
		r.Exec = func(value float32) (passed bool, violation string) {
			if *r.Min < value {
				return true, ""
			}

			return false, "value out of range"
		}

		return nil
	}

	if r.Max != nil {
		r.Exec = func(value float32) (passed bool, violation string) {
			if *r.Max > value {
				return true, ""
			}

			return false, "value out of range"
		}

		return nil
	}

	return nil
}
