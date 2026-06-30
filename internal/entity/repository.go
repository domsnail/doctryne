package entity

import (
	"net/url"
	"time"
)

type Repository struct {
	Name          string // github full repository name
	Description   string
	DefaultBranch string
	Homepage      string

	Owner        *Developer
	Organization *Organization

	Contributors []*Developer
	Maintainers  []*Developer

	// Only set from .git directory inspection, see at internal/service/git_service/service.go
	DeveloperCommitStats DeveloperCommitStats
	CommitStats          *CommitStats
	Commits              []*Commit

	Language string
	Size     uint64

	IsArchived bool
	IsDisabled bool
	IsFork     bool
	ForksCount uint64

	NetworkCount     uint64
	OpenIssuesCount  uint64 // including open pull requests
	StargazersCount  uint64
	SubscribersCount uint64

	License string

	GithubID int64
	GitURL   *url.URL

	CreatedAt time.Time
	UpdatedAt time.Time
	PushedAt  time.Time
}
