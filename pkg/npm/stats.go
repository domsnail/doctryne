package npm

type Stats struct {
	Package string `json:"package"`

	Downloads uint64 `json:"downloads"`

	Start string `json:"start"`
	End   string `json:"end"`
}
