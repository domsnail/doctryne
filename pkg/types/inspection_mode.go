package types

type InspectionMode string

const (
	// InspectionMode_Direct inspecting only the target entity
	InspectionMode_Direct InspectionMode = "direct"

	// InspectionMode_Shallow inspecting the target + direct related entities
	InspectionMode_Shallow InspectionMode = "shallow"

	// InspectionMode_Deep inspecting the target + all related entities recursively
	InspectionMode_Deep InspectionMode = "deep"
)
