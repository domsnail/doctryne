package types

type EvidenceOverride int

const (
	EvidenceOverride_Unspecified EvidenceOverride = iota

	EvidenceOverride_Affected
	EvidenceOverride_NotAffected
)
