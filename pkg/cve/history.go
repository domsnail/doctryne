package cve

import "time"

type HistoryQueryResponse struct {
	Format    string    `json:"format"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`

	ResultsPerPage uint32 `json:"resultsPerPage"`
	StartIndex     uint32 `json:"startIndex"`
	TotalResults   uint32 `json:"totalResults"`

	CveChanges []struct {
		Change `json:"change"`
	} `json:"cveChanges"`
}

type Change struct {
	CveID       string `json:"cveId"`
	CveChangeId string `json:"cveChangeId"`

	EventName        string `json:"eventName"`
	SourceIdentifier string `json:"sourceIdentifier"`

	Details []ChangeDetails `json:"details"`

	Created time.Time `json:"created"`
}

type ChangeDetails struct {
	Action   string `json:"action"`
	Type     string `json:"type"`
	NewValue string `json:"newValue"`
}
