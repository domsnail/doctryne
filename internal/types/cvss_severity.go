package types

type CvssSeverity string

const (
	Severity_Critical CvssSeverity = "critical"
	Severity_High     CvssSeverity = "high"
	Severity_Medium   CvssSeverity = "medium"
	Severity_Low      CvssSeverity = "low"

	Severity_None CvssSeverity = "none"
)
