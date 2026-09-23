package entity

import (
	"fmt"
	"net/url"
	"strconv"
)

type QueryFilter struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func (q QueryFilter) FromQuery(params url.Values) error {
	var err error
	if params.Has("limit") {
		q.Limit, err = strconv.Atoi(params.Get("limit"))
		if err != nil {
			return fmt.Errorf("invalid limit field: %w", err)
		}
	} else {
		q.Limit = 10
	}

	if params.Has("offset") {
		q.Offset, err = strconv.Atoi(params.Get("offset"))
		if err != nil {
			return fmt.Errorf("invalid offset field: %w", err)
		}
	}

	return nil
}

type Sorting struct {
	Field string `json:"field"`
	Order string `json:"order"`
}
