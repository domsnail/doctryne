package types

type ScanType string

const (
	ScanType_Unspecified ScanType = "unspecified"

	ScanType_Binary   ScanType = "binary_file"
	ScanType_FilePath ScanType = "file_path"
	ScanType_URL      ScanType = "url"
)

var ScanTypes = []ScanType{
	ScanType_Unspecified,
	ScanType_Binary,
	ScanType_FilePath,
	ScanType_URL,
}
