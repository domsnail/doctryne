package types

type ScanType string

const (
	ScanType_Unspecified ScanType = "unspecified"

	ScanType_Binary   ScanType = "binary_file"
	ScanType_FilePath ScanType = "file_path"
	ScanType_DirPath  ScanType = "dir_path"
	ScanType_URL      ScanType = "url"
)

var ScanTypes = []ScanType{
	ScanType_Unspecified,
	ScanType_FilePath,
	ScanType_DirPath,
	ScanType_Binary,
	ScanType_URL,
}

var ScanTypes_Enums = map[int32]ScanType{
	0: ScanType_Unspecified,
	1: ScanType_FilePath,
	2: ScanType_DirPath,
	3: ScanType_Binary,
	4: ScanType_URL,
}
