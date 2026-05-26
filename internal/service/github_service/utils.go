package github_service

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/google/go-github/v87/github"
)

func repositoryFromURL(link *url.URL) (owner, name string, err error) {
	slog.Debug("trying to determine github repository owner/name...",
		slog.String("repository_url", link.Redacted()),
	)

	path := strings.TrimSuffix(strings.Trim(link.EscapedPath(), "/"), ".git")

	switch link.Hostname() {
	case "github.com":
		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("repository url not supported: '%s'", link.String())
		}

		return parts[0], parts[1], nil
	default:
		return "", "", fmt.Errorf("unsupported git hostname '%s'", link.Hostname())
	}
}

func repositoriesToEntity(g []*github.Repository) []*entity.Repository {
	var repos = make([]*entity.Repository, len(g))
	for i, c := range g {
		repos[i] = repositoryToEntity(c)
	}

	return repos
}

func repositoryToEntity(g *github.Repository) *entity.Repository {
	r := entity.Repository{
		Name:             strings.TrimSpace(strings.ToLower(g.GetFullName())),
		Description:      g.GetDescription(),
		DefaultBranch:    g.GetDefaultBranch(),
		Homepage:         g.GetHomepage(),
		Owner:            userToEntity(g.GetOwner()),
		Organization:     organizationToEntity(g.GetOrganization()),
		Language:         g.GetLanguage(),
		Size:             uint64(g.GetSize()),
		IsArchived:       g.GetArchived(),
		IsDisabled:       g.GetDisabled(),
		IsFork:           g.GetFork(),
		ForksCount:       uint64(g.GetForksCount()),
		NetworkCount:     uint64(g.GetNetworkCount()),
		OpenIssuesCount:  uint64(g.GetOpenIssuesCount()),
		StargazersCount:  uint64(g.GetStargazersCount()),
		SubscribersCount: uint64(g.GetSubscribersCount()),
		GithubID:         g.GetID(),
	}

	if g.GitURL != nil {
		var err error
		r.GitURL, err = url.Parse(g.GetGitURL())
		if err != nil {
			slog.Warn(fmt.Sprintf("failed to parse github url: %s", err.Error()))
		}
	}

	if g.License != nil {
		r.License = g.GetLicense().GetName()
	}

	if g.CreatedAt != nil {
		r.CreatedAt = g.CreatedAt.Time
	}

	if g.UpdatedAt != nil {
		r.UpdatedAt = g.UpdatedAt.Time
	}

	if g.PushedAt != nil {
		r.PushedAt = g.PushedAt.Time
	}

	return &r
}

func usersToEntity(g []*github.User) []*entity.Developer {
	var dev = make([]*entity.Developer, len(g))
	for i, c := range g {
		dev[i] = userToEntity(c)
	}

	return dev
}

func userToEntity(g *github.User) *entity.Developer {
	if g == nil {
		return nil
	}

	p := entity.Developer{
		GithubID:          g.ID,
		Name:              g.GetName(),
		Username:          strings.ToLower(g.GetLogin()),
		TwitterUsername:   g.GetTwitterUsername(),
		Location:          g.GetLocation(),
		Company:           g.GetCompany(),
		Blog:              g.GetBlog(),
		Bio:               g.GetBio(),
		IsHireable:        g.GetHireable(),
		IsSiteAdmin:       g.GetSiteAdmin(),
		FollowersCount:    uint64(g.GetFollowers()),
		FollowingCount:    uint64(g.GetFollowing()),
		PublicReposCount:  uint64(g.GetPublicRepos()),
		PrivateReposCount: uint64(g.GetOwnedPrivateRepos()),
	}

	if g.UserViewType != nil {
		p.IsPrivate = *g.UserViewType != "public"
	}

	if g.Email != nil {
		p.Emails = []string{g.GetEmail()}
	}

	if g.CreatedAt != nil {
		p.CreatedAt = g.CreatedAt.Time
	}

	if g.UpdatedAt != nil {
		p.UpdatedAt = g.UpdatedAt.Time
	}

	if g.SuspendedAt != nil {
		p.SuspendedAt = &g.SuspendedAt.Time
	}

	return &p
}

func contributorsToEntity(g []*github.Contributor) []*entity.Developer {
	var dev = make([]*entity.Developer, len(g))
	for i, c := range g {
		dev[i] = contributorToEntity(c)
	}

	return dev
}

func contributorToEntity(g *github.Contributor) *entity.Developer {
	if g == nil {
		return nil
	}

	p := entity.Developer{
		GithubID:    g.ID,
		Name:        g.GetName(),
		Username:    strings.ToLower(g.GetLogin()),
		IsSiteAdmin: g.GetSiteAdmin(),
	}

	if g.Email != nil {
		p.Emails = []string{g.GetEmail()}
	}

	return &p
}

func organizationsToEntity(g []*github.Organization) []*entity.Organization {
	var dev = make([]*entity.Organization, len(g))
	for i, c := range g {
		dev[i] = organizationToEntity(c)
	}

	return dev
}

func organizationToEntity(g *github.Organization) *entity.Organization {
	if g == nil {
		return nil
	}

	p := entity.Organization{
		GithubID:           g.ID,
		Name:               g.GetName(),
		Username:           strings.ToLower(g.GetLogin()),
		Description:        g.GetDescription(),
		TwitterUsername:    g.GetTwitterUsername(),
		Location:           g.GetLocation(),
		Company:            g.GetCompany(),
		Blog:               g.GetBlog(),
		IsVerified:         g.GetIsVerified(),
		FollowersCount:     uint64(g.GetFollowers()),
		FollowingCount:     uint64(g.GetFollowing()),
		CollaboratorsCount: uint64(g.GetCollaborators()),
		PublicReposCount:   uint64(g.GetPublicRepos()),
		PrivateReposCount:  uint64(g.GetOwnedPrivateRepos()),
	}

	if g.Email != nil {
		p.Emails = []string{g.GetEmail()}
	}

	if g.BillingEmail != nil {
		p.Emails = []string{g.GetBillingEmail()}
	}

	if g.CreatedAt != nil {
		p.CreatedAt = g.CreatedAt.Time
	}

	if g.UpdatedAt != nil {
		p.UpdatedAt = g.UpdatedAt.Time
	}

	if g.ArchivedAt != nil {
		p.ArchivedAt = &g.ArchivedAt.Time
	}

	return &p
}
