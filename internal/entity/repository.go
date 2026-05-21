package entity

import (
	"net/url"
	"time"
)

type Repository struct {
	Name        string // github full repository name
	Description string

	DefaultBranch string
	Homepage      string
	Owner         *Person
	Organization  *Organization

	Language string
	Size     uint64

	IsArchived bool
	IsDisabled bool
	IsFork     bool
	ForksCount uint64

	NetworkCount     uint64
	OpenIssuesCount  uint64
	StargazersCount  uint64
	SubscribersCount uint64

	License string

	GithubID int64
	GitURL   *url.URL

	CreatedAt time.Time
	UpdatedAt time.Time
	PushedAt  time.Time
}
