package entity

import "github.com/domsnail/doctryne/internal/types"

type Evidence struct {
	Ratio float64 // positive numbers increases True Positive probability

	Comment   string
	Reference string
}

type VersionMatchEvidence struct {
	Status types.AffectedStatus

	InRange    bool
	Confidence float64
	Message    string
}

type VersionMatchEvidenceSet []VersionMatchEvidence

func (v VersionMatchEvidenceSet) Combine() VersionMatchEvidence {

	return VersionMatchEvidence{}
}
