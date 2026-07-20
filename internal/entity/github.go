package entity

import (
	"net/url"
	"time"
)

type GitHubRepositoryMetadata struct {
	ID     int64
	NodeID string

	Name          string // full repository name
	Description   string
	DefaultBranch string
	Homepage      string
	License       string
	Language      string

	IsArchived bool
	IsDisabled bool
	IsFork     bool
	ForksCount uint64

	Languages map[string]int
	Size      uint64

	NetworkCount     uint64
	OpenIssuesCount  uint64 // including open pull requests
	SubscribersCount uint64 // or watchers

	StargazersCount uint64 // stars

	Owner *Developer
	Org   *Organization

	// must be queried separately with another api call, see GetRepositoryContributors
	Contributors []*Developer
	Issues       []*GithubIssue

	GitURL *url.URL

	CreatedAt time.Time
	UpdatedAt time.Time
	PushedAt  time.Time
}
