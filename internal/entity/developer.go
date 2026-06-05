package entity

import "time"

type Developer struct {
	Name     string
	Username string
	Emails   []string

	TwitterUsername string

	Location string
	Company  string
	Blog     string
	Bio      string

	IsPrivate   bool
	IsHireable  bool
	IsSiteAdmin bool

	FollowersCount uint64
	FollowingCount uint64

	PublicReposCount  uint64
	PrivateReposCount uint64

	GithubID *int64

	LatestActivity []*Activity
	Repositories   []*Repository

	CreatedAt   time.Time
	UpdatedAt   time.Time
	SuspendedAt *time.Time
}
