package entity

import (
	"net/url"
	"time"
)

type Repository struct {
	Name        string // github full repository name
	Description string
	License     string

	Owner        *Developer
	Organization *Organization

	Contributors []*Developer
	Maintainers  []*Developer

	// Only set from .git directory inspection, see at internal/service/git_service/service.go
	GitURL               *url.URL
	DeveloperCommitStats DeveloperCommitStats
	CommitStats          *CommitStats
	Commits              []*Commit

	Language string
	Size     uint64

	GithubID       uint64
	GithubMetadata *GitHubRepositoryMetadata

	CreatedAt time.Time
	UpdatedAt time.Time
	PushedAt  time.Time
}
