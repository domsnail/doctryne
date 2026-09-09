package cve_v52

type Affected struct {
	Product string `json:"product"`
	Vendor  string `json:"vendor"`

	PackageName   string   `json:"packageName"`
	CollectionURL string   `json:"collectionURL" validated:"omitempty,url"`
	CPEs          []string `json:"cpes"`

	Platforms []string `json:"platforms" validated:"unique"`

	RepositoryURL string `json:"repo" validated:"omitempty,url"`

	DefaultStats AffectedStatus `json:"defaultStatus"`

	Versions []Version `json:"versions"`
}

type Version struct {
	Version     string `json:"version" validated:"omitempty,min=1"`
	VersionType string `json:"versionType"`

	Status AffectedStatus `json:"status"`

	LessThan        string `json:"lessThan"`
	LessThanOrEqual string `json:"lessThanOrEqual"`

	Changes []struct {
		At     string `json:"at"`
		Status string `json:"status"`
	}
}

type AffectedStatus string

const (
	AffectedStatus_Unknown    AffectedStatus = "unknown"
	AffectedStatus_Unaffected AffectedStatus = "unaffected"
	AffectedStatus_Affected   AffectedStatus = "affected"
)
