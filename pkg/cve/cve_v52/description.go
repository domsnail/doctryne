package cve_v52

// Description
// ref: https://cveproject.github.io/cve-schema/schema/docs/#oneOf_i0_containers_cna_descriptions_items
type Description struct {
	Lang  string `json:"lang" `
	Value string `json:"value" `

	SupportingMedia []SupportingMedia `json:"supportingMedia,omitempty"`
}

type SupportingMedia struct {
	CWEID string `json:"cweId,omitempty" `

	Type   string `json:"type" ` // RFC2046 compliant IANA Media type
	Base64 bool   `json:"base64"`

	// Supporting media content, up to 16K. If base64 is true, this field stores base64 encoded data.
	Value string `json:"value" `
}
