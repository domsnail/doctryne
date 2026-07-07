package entity

import (
	"net/url"
	"time"
)

type GitHubRepositoryMetadata struct {
	ID int64

	Name          string // full repository name
	Description   string
	DefaultBranch string
	Homepage      string
	License       string

	IsArchived bool
	IsDisabled bool
	IsFork     bool
	ForksCount uint64

	Language string
	Size     uint64

	NetworkCount     uint64
	OpenIssuesCount  uint64 // including open pull requests
	StargazersCount  uint64
	SubscribersCount uint64

	Owner *GithubDeveloperMetadata
	Org   *GithubOrganizationMetadata

	// must be queried separately with another api call, see GetRepositoryContributors
	Contributors []*GithubDeveloperMetadata

	GitURL *url.URL

	CreatedAt time.Time
	UpdatedAt time.Time
	PushedAt  time.Time
}
