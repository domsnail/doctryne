package entity

import (
	"net/url"
	"time"
)

type Repository struct {
	Name string

	// Usually is set from registry data
	GitURL *url.URL

	// Only set from .git directory inspection, see at internal/service/git_service/service.go
	DeveloperCommitStats DeveloperCommitStats
	Commiters            []*Developer
	CommitStats          *CommitStats
	Commits              []*Commit

	// Only set from GitHub API
	GithubID       int64
	GithubMetadata *GitHubRepositoryMetadata

	CreatedAt time.Time
	UpdatedAt time.Time
	PushedAt  time.Time
}
