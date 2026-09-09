package cve_v52

// Record CVE
// ref: https://nvd.nist.gov/developers/vulnerabilities
// ref: https://cveproject.github.io/cve-schema/schema/docs
// ref: https://csrc.nist.gov/schema/nvd/api/2.0/cve_api_json_2.0.schema
type Record struct {
	DataType    string `json:"dataType"`
	DataVersion string `json:"dataVersion"`

	CveMetadata Metadata `json:"cveMetadata" `

	Containers struct {
		CNA CNA   `json:"cna" `
		ADP []ADP `json:"adp"`
	} `json:"containers"`
}
