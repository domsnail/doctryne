package entity

import "github.com/domsnail/doctryne/internal/types"

type VulnerabilityFindings []*VulnerabilityFinding

type VulnerabilityFinding struct {
	Match    *VulnerabilityMatch           `json:"match,omitempty"`
	Evidence *VulnerabilityFindingEvidence `json:"evidence,omitempty"`
}

type VulnerabilityFindingEvidence struct {
	Method     types.MatcherMethod `json:"method"`
	Confidence float32             `json:"confidence"`
}

type VulnerabilityFindingOptions struct {
	Purl string `json:"purl,omitempty"`

	Ecosystem types.Ecosystem `json:"ecosystem,omitempty"`
	Name      string          `json:"name,omitempty"`
	Version   string          `json:"version,omitempty"`

	// show only ValidatedOnly vulnerabilities, in statuses: analyzed, modified
	ValidatedOnly bool `json:"validated_only,omitempty"`

	// also show Withdrawn vulnerabilities
	Withdrawn bool `json:"withdrawn,omitempty"`

	MinConfidence float32 `json:"min_confidence"`
	ShowEvidence  bool    `json:"show_evidence"`
}
