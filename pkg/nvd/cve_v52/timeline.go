package cve_v52

type Timeline struct {
	Lang  string `json:"lang" `
	Value string `json:"value" `

	Time *Timestamp `json:"time"`
}
