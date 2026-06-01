package types

type ManifestType string

const (
	ManifestType_Unspecified ManifestType = "unspecified"

	ManifestType_Go_Mod = "go.mod"
	ManifestType_Go_Sum = "go.sum"

	ManifestType_Package_Json = "package.json"
	ManifestType_Package_Lock = "package-lock.json"

	ManifestType_CycloneDX = "cyclonedx.json"
)

var ManifestTypes = []ManifestType{
	ManifestType_Unspecified,
	//
	ManifestType_Go_Mod,
	ManifestType_Go_Sum,
	//
	ManifestType_Package_Json,
	ManifestType_Package_Lock,
	//
	ManifestType_CycloneDX,
}
