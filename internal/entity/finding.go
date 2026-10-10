package entity

import "github.com/domsnail/doctryne/internal/types"

type VulnerabilityFindings []*VulnerabilityFinding

type VulnerabilityFinding struct {
	Vulnerability *VulnerabilityMatch           `json:"vulnerability,omitempty"`
	Evidence      *VulnerabilityFindingEvidence `json:"evidence,omitempty"`
}

type VulnerabilityFindingEvidence struct {
	Method     types.MatcherMethod `json:"method"`
	Confidence float32             `json:"confidence"`
}

type VulnerabilityFindingOptions struct {
	ValidatedOnly bool    // show only ValidatedOnly vulnerabilities, in statuses: analyzed, modified
	MinConfidence float32 `json:"min_confidence"`
}
