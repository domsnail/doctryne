package cve_v52

type Metrics struct {
	CvssV40 CVSSv40 `json:"cvssV4_0,omitempty"`
	CvssV31 CVSSv31 `json:"cvssV3_1,omitempty"`
	CvssV30 CVSSv30 `json:"cvssV3_0,omitempty"`
	CvssV20 CVSSv20 `json:"cvssV2_0,omitempty"`

	Other struct {
		// Name of the non-standard impact metrics format used.
		Type string `json:"type"`

		// not covered by CVE:
		// https://cveproject.github.io/cve-schema/schema/docs/#oneOf_i0_containers_cna_metrics_items_other_content
		Content interface{} `json:"content"`
	} `json:"other"`
}

type OtherMetrics struct {
	Id        string    `json:"id"`
	Timestamp Timestamp `json:"timestamp"`

	Options []struct {
		Exploitation    string `json:"Exploitation,omitempty"`
		Automatable     string `json:"Automatable,omitempty"`
		TechnicalImpact string `json:"Technical Impact,omitempty"`
	} `json:"options"`

	Role    string `json:"role"`
	Version string `json:"version"`
}
