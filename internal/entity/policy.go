package entity

type Policy struct {
	MaxTotalViolations uint64
	MaxCVSSScore       float32
	MaxSeverity        float32

	DisallowedCWEIds []string
}
