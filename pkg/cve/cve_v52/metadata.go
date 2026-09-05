package cve_v52

import (
	"github.com/google/uuid"
)

type Metadata struct {
	CveID  string `json:"cveId"`
	State  State  `json:"state"`
	Serial int    `json:"serial"`

	AssignerOrgId     *uuid.UUID `json:"assignerOrgId"`
	AssignerShortName string     `json:"assignerShortName"`

	RequesterUserId *uuid.UUID `json:"requesterUserId"`

	DateReserved  *Timestamp `json:"dateReserved"`
	DateRejected  *Timestamp `json:"dateRejected"`
	DatePublished *Timestamp `json:"datePublished"`
	DateUpdated   *Timestamp `json:"dateUpdated,omitempty"`
}

type State string

const (
	State_Published State = "PUBLISHED"
	State_Rejected  State = "REJECTED"
)
