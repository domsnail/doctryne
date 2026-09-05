package cve_v52

type ADP struct {
	Title            string           `json:"title"`
	ProviderMetadata ProviderMetadata `json:"providerMetadata"`

	ProblemTypes []ProblemType `json:"problemTypes"`
	Metrics      []Metrics     `json:"metrics"`

	References []Reference `json:"references"`

	Affected []Affected `json:"affected" `
	Impacts  []Impact   `json:"impacts"`

	Descriptions   []Description `json:"descriptions" `
	Configurations []Description `json:"configurations" `
	Workarounds    []Description `json:"workarounds" `
	Solutions      []Description `json:"solutions" `
	Exploits       []Description `json:"exploits" `

	Timeline []Timeline `json:"timeline" `

	Tags []string `json:"tags"`

	Source  *Source       `json:"source,omitempty"`
	Credits []Description `json:"credits" `
}
