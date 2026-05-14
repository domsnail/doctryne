package types

type ScanType string

const (
	ScanType_Unspecified ScanType = "unspecified"

	ScanType_Files     ScanType = "files"
	ScanType_CycloneDX ScanType = "cyclonedx"
	ScanType_URL       ScanType = "url"
)

var ScanTypes = []ScanType{
	ScanType_Unspecified,
	ScanType_Files,
	ScanType_CycloneDX,
	ScanType_URL,
}
