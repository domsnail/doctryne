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
	ManifestType_Go_Mod,
	//
	ManifestType_Package_Json,
}

var LockfileTypes = []ManifestType{
	ManifestType_Go_Sum,
	//
	ManifestType_Package_Lock,
}

var ManifestType_Language = map[ManifestType]Language{
	ManifestType_Go_Mod:       Language_Golang,
	ManifestType_Go_Sum:       Language_Golang,
	ManifestType_Package_Json: Language_JavaScript,
	ManifestType_Package_Lock: Language_JavaScript,
}
