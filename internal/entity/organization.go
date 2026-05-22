package entity

import "time"

type Organization struct {
	Name        string
	Username    string
	Description string
	Emails      []string

	TwitterUsername string

	Location string
	Company  string
	Blog     string

	IsVerified bool

	FollowersCount     uint64
	FollowingCount     uint64
	CollaboratorsCount uint64

	PublicReposCount  uint64
	PrivateReposCount uint64

	GithubID *int64

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt *time.Time
}
