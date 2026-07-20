package entity

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"
)

type Organization struct {
	Name   string
	Login  string
	Emails []string

	GithubID       int64
	GithubMetadata *GithubOrganizationMetadata
}

func (org *Organization) Merge(other *Organization) error {
	if other == nil {
		return nil
	} else if !strings.EqualFold(org.Login, other.Login) {
		return fmt.Errorf("cannot merge organizations with different logins")
	}

	if other.GithubID != 0 {
		if org.GithubID != 0 && other.GithubID != org.GithubID {
			return fmt.Errorf("cannot merge organizations with different github profiles")
		}

		org.GithubID = other.GithubID

		if other.GithubMetadata != nil {
			org.GithubMetadata = other.GithubMetadata
		}
	}

	if len(other.Emails) > 0 {
		var emails = make(map[string]bool)

		for _, e := range append(other.Emails, org.Emails...) {
			emails[strings.ToLower(e)] = false
		}

		org.Emails = slices.Collect(maps.Keys(emails))
	}

	return nil
}

type GithubOrganizationMetadata struct {
	ID     int64
	NodeID string

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

func (md *GithubOrganizationMetadata) Organization() *Organization {
	return &Organization{
		Name:           md.Name,
		Login:          md.Login,
		Emails:         md.Emails,
		GithubID:       md.ID,
		GithubMetadata: md,
	}
}
