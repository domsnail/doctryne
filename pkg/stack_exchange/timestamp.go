package stack_exchange

import (
	"encoding/json"
	"time"
)

type UnixTimestamp struct {
	time.Time
}

func (u *UnixTimestamp) UnmarshalJSON(b []byte) error {
	var timestamp int64

	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}

	u.Time = time.Unix(timestamp, 0)
	return nil
}

func (u *UnixTimestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Time.Unix())
}
