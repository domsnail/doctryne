package cve_v52

type ProviderMetadata struct {
	OrgId     string `json:"orgId"`
	ShortName string `json:"shortName" `

	// DateUpdated can be set without timezone data
	DateUpdated *Timestamp `json:"dateUpdated,omitempty"`
}
