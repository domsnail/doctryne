package cve_v52

type Impact struct {
	CAPEC        string        `json:"capecId" `
	Descriptions []Description `json:"descriptions" `
}
