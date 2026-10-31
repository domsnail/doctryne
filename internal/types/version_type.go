package types

type VersionType string

const (
	VersionType_None    VersionType = ""
	VersionType_Unknown VersionType = "unknown"

	VersionType_Semver VersionType = "semver"
	VersionType_Custom VersionType = "custom"
	VersionType_Git    VersionType = "git"
)

func GetVersionType(v string) VersionType {
	return VersionType_Unknown
}
