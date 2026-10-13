package types

type AffectedStatus string

const (
	AffectedStatus_Unspecified AffectedStatus = "unspecified"

	AffectedStatus_Affected   AffectedStatus = "affected"
	AffectedStatus_Unaffected AffectedStatus = "unaffected"
	AffectedStatus_Unknown    AffectedStatus = "unknown"
)
