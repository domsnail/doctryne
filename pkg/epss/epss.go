package epss

import "time"

type EpssDatabase struct {
	Version string
	Date    time.Time
	Records []EpssRecord
}

type EpssRecord struct {
	CveID      string
	Epss       float32
	Percentile float32
}
