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
		return "", "", fmt.Errorf("unsupported git vcs hostname '%s'", link.Hostname())
	}
}

func repositoriesToMetadata(g []*github.Repository) []*entity.GitHubRepositoryMetadata {
	var repos = make([]*entity.GitHubRepositoryMetadata, len(g))
	for i, c := range g {
		repos[i] = repositoryToMetadata(c)
	}

	return repos
}

func repositoryToMetadata(g *github.Repository) *entity.GitHubRepositoryMetadata {
	md := entity.GitHubRepositoryMetadata{
		ID:               g.GetID(),
		Name:             g.GetName(),
		Description:      g.GetDescription(),
		DefaultBranch:    g.GetDefaultBranch(),
		Homepage:         g.GetHomepage(),
		IsArchived:       g.GetArchived(),
		IsDisabled:       g.GetDisabled(),
		IsFork:           g.GetFork(),
		Language:         g.GetLanguage(),
		Size:             uint64(g.GetSize()),
		Owner:            userToMetadata(g.GetOwner()),
		Org:              organizationToMetadata(g.GetOrganization()),
		ForksCount:       uint64(g.GetForksCount()),
		NetworkCount:     uint64(g.GetNetworkCount()),
		OpenIssuesCount:  uint64(g.GetOpenIssuesCount()),
		StargazersCount:  uint64(g.GetStargazersCount()),
		SubscribersCount: uint64(g.GetSubscribersCount()),
	}

	if g.GitURL != nil {
		var err error
		md.GitURL, err = url.Parse(g.GetGitURL())
		if err != nil {
			slog.Warn(fmt.Sprintf("failed to parse github url: %s", err.Error()))
		}
	}

	if g.License != nil {
		md.License = g.GetLicense().GetName()
	}

	if g.CreatedAt != nil {
		md.CreatedAt = g.CreatedAt.Time
	}

	if g.UpdatedAt != nil {
		md.UpdatedAt = g.UpdatedAt.Time
	}

	if g.PushedAt != nil {
		md.PushedAt = g.PushedAt.Time
	}

	return &md
}

func usersToMetadata(g []*github.User) []*entity.GithubDeveloperMetadata {
	var dev = make([]*entity.GithubDeveloperMetadata, len(g))
	for i, c := range g {
		dev[i] = userToMetadata(c)
	}

	return dev
}

func userToMetadata(g *github.User) *entity.GithubDeveloperMetadata {
	if g == nil {
		return nil
	}

	md := entity.GithubDeveloperMetadata{
		ID:                g.GetID(),
		Username:          g.GetLogin(),
		Fullname:          g.GetName(),
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
		md.IsPrivate = *g.UserViewType != "public"
	}

	if g.Email != nil {
		md.Email = g.GetEmail()
	}

	if g.CreatedAt != nil {
		md.CreatedAt = g.CreatedAt.Time
	}

	if g.UpdatedAt != nil {
		md.UpdatedAt = g.UpdatedAt.Time
	}

	if g.SuspendedAt != nil {
		md.SuspendedAt = &g.SuspendedAt.Time
	}

	return &md
}

func contributorsToMetadata(g []*github.Contributor) []*entity.GithubDeveloperMetadata {
	var dev = make([]*entity.GithubDeveloperMetadata, len(g))
	for i, c := range g {
		dev[i] = contributorToEntity(c)
	}

	return dev
}

func contributorToEntity(g *github.Contributor) *entity.GithubDeveloperMetadata {
	if g == nil {
		return nil
	}

	p := entity.GithubDeveloperMetadata{
		ID:          g.GetID(),
		Fullname:    g.GetName(),
		Username:    strings.ToLower(g.GetLogin()),
		IsSiteAdmin: g.GetSiteAdmin(),
	}

	if g.Email != nil {
		p.Email = g.GetEmail()
	}

	return &p
}

func organizationsToEntity(g []*github.Organization) []*entity.GithubOrganizationMetadata {
	var dev = make([]*entity.GithubOrganizationMetadata, len(g))
	for i, c := range g {
		dev[i] = organizationToMetadata(c)
	}

	return dev
}

func organizationToMetadata(g *github.Organization) *entity.GithubOrganizationMetadata {
	if g == nil {
		return nil
	}

	p := entity.GithubOrganizationMetadata{
		ID:                 g.GetID(),
		Name:               g.GetName(),
		Login:              strings.ToLower(g.GetLogin()),
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

func issuesToEntity(issues []*github.Issue) []*entity.Issue {
	var iss = make([]*entity.Issue, len(issues))

	for i, c := range issues {
		iss[i] = issueToMetadata(c)
	}

	return iss
}

func issueToMetadata(issue *github.Issue) *entity.Issue {
	return &entity.Issue{}
}
