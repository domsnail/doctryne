package entity

import (
	"time"

	"github.com/domsnail/doctryne/internal/types"
)

type VulnerabilityFindings []*VulnerabilityFinding

type VulnerabilityFinding struct {
	Match *VulnerabilityMatch `json:"match,omitempty"`
}

type VulnerabilityFindingOptions struct {
	Purl string `json:"purl,omitempty"`

	Ecosystem types.Ecosystem `json:"ecosystem,omitempty"`
	Name      string          `json:"name,omitempty"`
	Version   string          `json:"version,omitempty"`

	// show only vulnerabilities that exist on a specified Date
	Date *time.Time `json:"date,omitempty"`

	// show only ValidatedOnly vulnerabilities, in statuses: analyzed, modified
	ValidatedOnly bool `json:"validated_only,omitempty"`

	// also show Withdrawn vulnerabilities
	Withdrawn bool `json:"withdrawn,omitempty"`

	MinConfidence  float64 `json:"min_confidence"`
	ShowEvidence   bool    `json:"show_evidence"`
	ShowUnaffected bool    `json:"show_unaffected"`
}
