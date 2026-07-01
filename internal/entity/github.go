package entity

type GitHubRepositoryMetadata struct {
	ID uint64

	Name          string // full repository name
	Description   string
	DefaultBranch string
	Homepage      string

	IsArchived bool
	IsDisabled bool
	IsFork     bool
	ForksCount uint64

	NetworkCount     uint64
	OpenIssuesCount  uint64 // including open pull requests
	StargazersCount  uint64
	SubscribersCount uint64
}
