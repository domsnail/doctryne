package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/google/go-github/v87/github"
)

const itemsPerPage = 100

type GithubServiceImpl struct {
	c *github.Client
}

type GithubServiceOpts struct {
	Timeout  time.Duration
	ProxyURL *url.URL

	AccessToken string

	LatestActivityDepth int
}

func NewGithubServiceImpl(opts GithubServiceOpts) (*GithubServiceImpl, error) {
	slog.Debug("initializing github client...",
		slog.Bool("using_access_token", opts.AccessToken != ""),
	)

	if opts.Timeout == 0 {
		return nil, errors.New("timeout is required")
	}

	var (
		transport = http.DefaultTransport
		client    *github.Client
		err       error
	)

	if opts.ProxyURL != nil {
		slog.Debug("using proxy for github client", slog.String("proxy_url", opts.ProxyURL.Redacted()))
		transport = &http.Transport{
			Proxy: http.ProxyURL(opts.ProxyURL),
		}
	}

	var githubClientOpts = []github.ClientOptionsFunc{
		github.WithTimeout(opts.Timeout),
		github.WithTransport(transport),
	}

	if opts.AccessToken != "" {
		githubClientOpts = append(githubClientOpts, github.WithAuthToken(opts.AccessToken))
	}

	client, err = github.NewClient(githubClientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize github client: %w", err)
	}

	return &GithubServiceImpl{c: client}, nil
}

func (service GithubServiceImpl) Ping(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	slog.DebugContext(ctx, "pinging github...")

	me, _, err := service.c.Users.Get(ctx, "")
	if err != nil {
		slog.DebugContext(ctx, "failed to fetch github user",
			slog.String("error", err.Error()),
		)

		return fmt.Errorf("failed to ping github: %w", err)
	}

	slog.DebugContext(ctx, "successfully authenticated as github user",
		slog.Int64("user_id", me.GetID()),
		slog.String("username", me.GetLogin()),
		slog.String("profile_url", me.GetHTMLURL()),
		slog.String("email", me.GetEmail()),
		slog.String("plan", me.GetPlan().GetName()),
		slog.Bool("2fa_enabled", me.GetTwoFactorAuthentication()),
	)

	slog.InfoContext(ctx, "provided active github access token",
		slog.String("username", me.GetLogin()),
	)

	return nil
}

func (service GithubServiceImpl) GetRepositoryByName(ctx context.Context, owner, name string) (*entity.Repository, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	} else if owner == "" || name == "" {
		return nil, fmt.Errorf("repository owner and name are required")
	}

	slog.DebugContext(ctx, "fetching github repository...",
		slog.String("repository_path", fmt.Sprintf("%s/%s", strings.ToLower(owner), strings.ToLower(name))),
	)

	githubRepo, _, err := service.c.Repositories.Get(ctx, strings.ToLower(owner), strings.ToLower(name))
	if err != nil {
		slog.DebugContext(ctx, "failed to fetch repository",
			slog.String("repository_path", fmt.Sprintf("%s/%s", strings.ToLower(owner), strings.ToLower(name))),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	return repositoryToEntity(githubRepo), nil
}

func (service GithubServiceImpl) GetRepositoryByURL(ctx context.Context, link *url.URL) (*entity.Repository, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	owner, name, err := repositoryFromURL(link)
	if err != nil {
		slog.WarnContext(ctx, "failed to determine repository owner/name",
			slog.String("repository_url", link.Redacted()),
			slog.String("error", err.Error()),
		)
	}

	return service.GetRepositoryByName(ctx, owner, name)
}

func (service GithubServiceImpl) GetUserByUsername(ctx context.Context, username string) (*entity.Owner, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	} else if username == "" {
		return nil, fmt.Errorf("github username is required")
	}

	slog.DebugContext(ctx, "fetching github user...",
		slog.String("username", username),
	)

	user, _, err := service.c.Users.Get(ctx, username)
	if err != nil {
		slog.DebugContext(ctx, "failed to fetch github user",
			slog.String("username", username),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("failed to fetch github user: %w", err)
	}

	return userToEntity(user), nil
}

// GetUserActivity returns up to 300 events (max past 90 days)
func (service GithubServiceImpl) GetUserActivity(ctx context.Context, username string, depth time.Duration) (*entity.Activity, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	} else if username == "" {
		return nil, fmt.Errorf("github username is required")
	}

	slog.DebugContext(ctx, "fetching github user activity...",
		slog.String("username", username),
		slog.Duration("depth", depth),
	)

	_, _, err := service.c.Activity.ListEventsPerformedByUser(ctx, username, true, &github.ListOptions{
		PerPage: itemsPerPage,
	})

	return nil, err
}

func (service GithubServiceImpl) GetOrganizationByName(ctx context.Context, name string) (*entity.Organization, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	} else if name == "" {
		return nil, fmt.Errorf("github organization name is required")
	}

	slog.DebugContext(ctx, "fetching github organization...",
		slog.String("name", name),
	)

	organization, _, err := service.c.Organizations.Get(ctx, name)
	if err != nil {
		slog.DebugContext(ctx, "failed to fetch github organization",
			slog.String("name", name),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("failed to fetch github organization: %w", err)
	}

	return organizationToEntity(organization), nil
}

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

func userToEntity(g *github.User) *entity.Owner {
	if g == nil {
		return nil
	}

	p := entity.Owner{
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
