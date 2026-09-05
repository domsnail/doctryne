package types

type DetectionMethod string

const (
	DetectionMethod_Exact   DetectionMethod = "exact"
	DetectionMethod_Partial DetectionMethod = "partial"
	DetectionMethod_Regex   DetectionMethod = "regex"
)
