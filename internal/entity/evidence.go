package entity

import (
	"strings"

	"github.com/domsnail/doctryne/internal/types"
)

type Evidence struct {
	Ratio float64 // positive numbers increases True Positive probability

	Comment   string
	Reference string
}

type VersionMatchEvidence struct {
	Status types.AffectedStatus `json:"status"`

	InRange    bool    `json:"in_range"`
	Confidence float64 `json:"confidence"`
	Message    string  `json:"message,omitempty"`
}

type VersionMatchEvidenceSet []VersionMatchEvidence

// Combine merges evidences into a single verdict. In-range evidences take precedence over
// out-of-range ones; conflicting in-range statuses reduce the resulting confidence.
func (v VersionMatchEvidenceSet) Combine() VersionMatchEvidence {
	var inRange, outOfRange VersionMatchEvidenceSet

	for _, e := range v {
		if e.Status == "" || e.Status == types.AffectedStatus_Unspecified {
			continue
		}

		e.Confidence = clampConfidence(e.Confidence)
		if e.Confidence == 0 {
			continue
		}

		if e.InRange {
			inRange = append(inRange, e)
		} else {
			outOfRange = append(outOfRange, e)
		}
	}

	if len(inRange) > 0 {
		return inRange.combineInRange()
	}

	if len(outOfRange) > 0 {
		return outOfRange.combineOutOfRange()
	}

	return VersionMatchEvidence{
		Status:  types.AffectedStatus_Unspecified,
		Message: v.messages(""),
	}
}

func (v VersionMatchEvidenceSet) combineInRange() VersionMatchEvidence {
	// or per status: independent evidences of the same status reinforce each other
	disbelief := make(map[types.AffectedStatus]float64)
	for _, e := range v {
		if _, ok := disbelief[e.Status]; !ok {
			disbelief[e.Status] = 1
		}

		disbelief[e.Status] *= 1 - e.Confidence
	}

	var best types.AffectedStatus
	bestScore, opposingScore := -1.0, 0.0

	for status, d := range disbelief {
		score := 1 - d
		if score > bestScore {
			opposingScore = 1 - (1-opposingScore)*(1-clampConfidence(bestScore))
			best, bestScore = status, score
		} else {
			opposingScore = 1 - (1-opposingScore)*(1-score)
		}
	}

	if opposingScore >= bestScore {
		return VersionMatchEvidence{
			Status:     types.AffectedStatus_Unknown,
			InRange:    true,
			Confidence: 0.5,
			Message:    "conflicting evidences: " + v.messages(""),
		}
	}

	return VersionMatchEvidence{
		Status:     best,
		InRange:    true,
		Confidence: bestScore * (1 - opposingScore),
		Message:    v.messages(best),
	}
}

func (v VersionMatchEvidenceSet) combineOutOfRange() VersionMatchEvidence {
	counts := make(map[types.AffectedStatus]int)
	var best types.AffectedStatus
	for _, e := range v {
		counts[e.Status]++
		if counts[e.Status] > counts[best] {
			best = e.Status
		}
	}

	confidence := 1.0
	for _, e := range v {
		if e.Status == best && e.Confidence < confidence {
			confidence = e.Confidence
		}
	}
	if len(counts) > 1 {
		confidence *= float64(counts[best]) / float64(len(v))
	}

	return VersionMatchEvidence{
		Status:     best,
		InRange:    false,
		Confidence: confidence,
		Message:    v.messages(best),
	}
}

func (v VersionMatchEvidenceSet) messages(status types.AffectedStatus) string {
	messages := make([]string, 0, len(v))
	for _, e := range v {
		if e.Message != "" && (status == "" || e.Status == status) {
			messages = append(messages, e.Message)
		}
	}

	return strings.Join(messages, "; ")
}

func clampConfidence(c float64) float64 {
	if c < 0 {
		return 0
	}

	if c > 1 {
		return 1
	}

	return c
}
