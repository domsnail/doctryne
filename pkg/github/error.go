package github

type ApiError struct {
	Status string `json:"status"`

	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url"`
}
