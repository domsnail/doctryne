package entity

import "time"

type Organization struct {
	Name string

	GithubID       *int64
	GithubMetadata *GithubOrganizationMetadata
}

type GithubOrganizationMetadata struct {
	ID int64

	Name        string
	Login       string
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

	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt *time.Time
}
