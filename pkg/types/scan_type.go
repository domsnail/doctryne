package types

type ScanType string

const (
	ScanType_Unspecified ScanType = "unspecified"

	ScanType_BinaryFile ScanType = "binary_file"
	ScanType_FilePath   ScanType = "file_path"
	ScanType_URL        ScanType = "url"
)

var ScanTypes = []ScanType{
	ScanType_Unspecified,
	ScanType_BinaryFile,
	ScanType_FilePath,
	ScanType_URL,
}
