package entity

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"
)

type Developer struct {
	Name     string
	Username string
	Emails   []string

	GithubID       int64
	GithubMetadata *GithubDeveloperProfile

	StackExchangeAccountID uint64
	StackExchangeProfile   *StackExchangeDeveloperProfile

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (d *Developer) String() string {
	return fmt.Sprintf("%s (@%s)", d.Name, d.Username)
}

func (d *Developer) IsEqual(other *Developer) bool {
	return strings.EqualFold(d.Username, other.Username) && strings.EqualFold(d.Name, other.Name)
}

// Merge adds all non-existing data to the original Developer, function
// returns error in case there are conflicting values
func (d *Developer) Merge(other *Developer) error {
	if other == nil {
		return nil
	} else if !d.IsEqual(other) {
		return fmt.Errorf("cannot merge developers with different names/logins")
	}

	if other.GithubID != 0 {
		if d.GithubID != 0 && other.GithubID != d.GithubID {
			return fmt.Errorf("cannot merge developers with different github profiles")
		}

		d.GithubID = other.GithubID

		if other.GithubMetadata != nil {
			d.GithubMetadata = other.GithubMetadata
		}
	}

	if len(other.Emails) > 0 {
		var emails = make(map[string]bool)

		for _, e := range append(other.Emails, d.Emails...) {
			emails[strings.ToLower(e)] = false
		}

		d.Emails = slices.Collect(maps.Keys(emails))
	}

	return nil
}

func (d *Developer) AddGithubProfile(profile *GithubDeveloperProfile) error {
	if profile == nil || profile.ID == 0 {
		return nil
	} else if d.GithubID != 0 && d.GithubID != profile.ID {
		return fmt.Errorf("cannot merge developers with different github profiles")
	}

	d.GithubMetadata = profile
	d.GithubID = profile.ID

	if profile.Email != "" {
		var emails = make(map[string]bool)

		for _, e := range append(d.Emails, profile.Email) {
			emails[strings.ToLower(e)] = false
		}

		d.Emails = slices.Collect(maps.Keys(emails))
	}

	return nil
}

type GithubDeveloperProfile struct {
	ID     int64
	NodeID string

	Username string
	Fullname string
	Email    string

	TwitterUsername string
	Location        string
	Company         string
	Blog            string
	Bio             string

	IsPrivate   bool
	IsHireable  bool
	IsSiteAdmin bool

	FollowersCount uint64
	FollowingCount uint64

	PublicReposCount  uint64
	PrivateReposCount uint64

	LatestActivity []*Activity
	Repositories   []*Repository

	CreatedAt   time.Time
	UpdatedAt   time.Time
	SuspendedAt *time.Time
}

func (profile *GithubDeveloperProfile) ToDeveloper() *Developer {
	d := Developer{
		Name:           profile.Fullname,
		Username:       profile.Username,
		GithubID:       profile.ID,
		GithubMetadata: profile,
	}

	if profile.Email != "" {
		d.Emails = []string{profile.Email}
	}

	return &d
}

type GithubStargazer struct {
	Username  string
	StarredAt *time.Time
}

type StackExchangeDeveloperProfile struct {
	UserID    uint64
	AccountID uint64

	DisplayName string
	WebsiteUrl  string
	AboutMe     string
	Location    string

	IsEmployee   bool
	IsRegistered bool

	Reputation uint32
	Badges     StackExchangeBadges

	CreatedAt    time.Time
	UpdatedAt    *time.Time
	LastAccessAt time.Time
	PenaltyTill  *time.Time
}

type StackExchangeBadges struct {
	Bronze uint32
	Silver uint32
	Gold   uint32
}
