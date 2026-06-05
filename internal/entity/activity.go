package entity

import "time"

type Activity struct {
	Action string
	Object string

	Timestamp time.Time
}
