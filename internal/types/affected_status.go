package types

type AffectedStatus string

const (
	AffectedStatus_Unspecified AffectedStatus = "unspecified"

	AffectedStatus_Affected   AffectedStatus = "affected"
	AffectedStatus_Unaffected AffectedStatus = "unaffected"
	AffectedStatus_Unknown    AffectedStatus = "unknown"

	// AffectedStatus_WillNotFix status means that a fix for an affected product version is not planned or not possible due to complexity, which may create additional risk.
	// ref: https://access.redhat.com/security/glossary#:~:text=Will%20not%20fix
	AffectedStatus_WillNotFix AffectedStatus = "will_not_fix"

	// AffectedStatus_FixDeferred is a deferred flaw status means that a fix for an affected product version is not guaranteed due to higher-priority development work.
	// ref: https://access.redhat.com/security/glossary#:~:text=Fix%20deferred
	AffectedStatus_FixDeferred AffectedStatus = "fix_deferred"

	AffectedStatus_EndOfLife AffectedStatus = "end_of_life"
)
