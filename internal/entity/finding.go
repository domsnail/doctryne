package entity

import "github.com/domsnail/doctryne/internal/types"

type VulnerabilityFindings []*VulnerabilityFinding

type VulnerabilityFinding struct {
	Vulnerability *Vulnerability                `json:"vulnerability,omitempty"`
	Evidence      *VulnerabilityFindingEvidence `json:"evidence,omitempty"`
}

type VulnerabilityFindingEvidence struct {
	Method     types.MatcherMethod
	Confidence float32
}

type VulnerabilityFindingOptions struct {
	ValidatedOnly bool // show only ValidatedOnly vulnerabilities, in statuses: analyzed, modified
	MinConfidence float32
}
