package entity

import "time"

type Person struct {
	Name   string
	Emails []string

	TwitterUsername string

	Location string
	Company  string
	Blog     string
	Bio      string

	IsHireable  bool
	IsSiteAdmin bool

	FollowersCount uint64
	FollowingCount uint64

	PublicReposCount  uint64
	PrivateReposCount uint64

	GithubID *int64

	CreatedAt   time.Time
	UpdatedAt   time.Time
	SuspendedAt time.Time
}
