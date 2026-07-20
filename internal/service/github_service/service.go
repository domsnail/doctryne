package github_service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/google/go-github/v87/github"

	http_smart_transport "github.com/domsnail/doctryne/pkg/http"
)

const (
	itemsPerPage = 100 // max batch size for github api
	maxPages     = 3

	defaultCacheTTL = 24 * time.Hour

	// github rate limits are different with or without PAT, but github client already have rate limit by default
	// ref: https://docs.github.com/ru/rest/using-the-rest-api/rate-limits-for-the-rest-api?apiVersion=2026-03-10
	defaultRateLimit_Period   = time.Minute * 60
	defaultRateLimit_MinDelay = 500 * time.Millisecond

	defaultRateLimit_MaxRequests = 60
	patRateLimit_MaxRequests     = 5000

	defaultRequestsDelay = 500 * time.Millisecond
)

type GithubServiceImpl struct {
	c *github.Client

	opts GithubServiceOpts
}

type GithubServiceOpts struct {
	AccessToken string

	LatestActivityPeriod time.Duration
	CacheTTL             time.Duration
}

func NewGithubServiceImpl(opts GithubServiceOpts) *GithubServiceImpl {
	slog.Debug("initializing github client...",
		slog.Bool("using_access_token", opts.AccessToken != ""),
	)

	if opts.LatestActivityPeriod <= 0 {
		slog.Debug("latest activity is not set or invalid, setting from global config")
		opts.LatestActivityPeriod = cfg.GlobalConfig.Scan.ActivityPeriod
	}

	if opts.CacheTTL == 0 {
		slog.Warn("no cache ttl set for github, setting to default value",
			slog.Duration("default_cache_ttl", defaultCacheTTL),
		)

		opts.CacheTTL = defaultCacheTTL
	}

	transportOpts := http_smart_transport.TransportOptions{
		BaseTransport: http.DefaultTransport,
		CachedMethods: []string{http.MethodGet},
		CacheTTL:      opts.CacheTTL,
		ThrottleOptions: &http_smart_transport.ThrottleOptions{
			RefreshPeriod: defaultRateLimit_Period,
			MaxRequests:   defaultRateLimit_MaxRequests,
			MinDelay:      defaultRateLimit_MinDelay,
		},
	}

	if opts.AccessToken == "" && cfg.GlobalConfig.Credentials.GithubApiKey != "" {
		slog.Debug("access token is not set, setting from global config")
		opts.AccessToken = cfg.GlobalConfig.Credentials.GithubApiKey
	}

	if opts.AccessToken != "" {
		slog.Debug("github access token is set, requests rate limit increased")
		transportOpts.ThrottleOptions.MaxRequests = patRateLimit_MaxRequests
	} else {
		slog.Warn("github access token is not set",
			slog.String("details", "please consider using github personal access token"),
		)
	}

	var (
		transport = http_smart_transport.NewSmartTransport(transportOpts)
		client    *github.Client
		err       error
	)

	var githubClientOpts = []github.ClientOptionsFunc{
		github.WithHTTPClient(&http.Client{
			Transport: transport,
			Timeout:   http.DefaultClient.Timeout,
		}),
	}

	if opts.AccessToken != "" {
		githubClientOpts = append(githubClientOpts, github.WithAuthToken(opts.AccessToken))
	}

	client, err = github.NewClient(githubClientOpts...)
	if err != nil {
		slog.Error("failed to initialize github client", slog.String("error", err.Error()))
	}

	slog.Info("initialized github client",
		slog.Bool("using_access_token", opts.AccessToken != ""),
		slog.Duration("latest_activity_period", opts.LatestActivityPeriod),
		slog.Duration("cache_ttl", opts.CacheTTL),
		slog.Duration("request_timeout", http.DefaultClient.Timeout),
		slog.Group("rate_limiting",
			slog.Duration("period", defaultRateLimit_Period),
			slog.Uint64("max_requests", transportOpts.ThrottleOptions.MaxRequests),
			slog.Duration("min_delay", defaultRateLimit_MinDelay),
		),
	)

	return &GithubServiceImpl{c: client, opts: opts}
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

// === GitHub Repository ===

func (service GithubServiceImpl) GetRepositoryByName(ctx context.Context, owner string, name string) (*entity.GitHubRepositoryMetadata, error) {
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

	return repositoryToMetadata(githubRepo), nil
}

func (service GithubServiceImpl) GetRepositoryByURL(ctx context.Context, link *url.URL) (*entity.GitHubRepositoryMetadata, error) {
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

func (service GithubServiceImpl) GetUserOwnedRepositories(ctx context.Context, username string) ([]*entity.GitHubRepositoryMetadata, error) {
	if username == "" {
		return nil, fmt.Errorf("github username is required")
	}

	var page = 1
	var repos []*entity.GitHubRepositoryMetadata

	for page <= maxPages {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		slog.DebugContext(ctx, "fetching github user repositories...",
			slog.String("username", username),
			slog.Int("items_per_page", itemsPerPage),
			slog.Int("page", page),
		)

		r, _, err := service.c.Repositories.ListByUser(ctx, username, &github.RepositoryListByUserOptions{
			Type:      "owner",
			Sort:      "updated",
			Direction: "desc",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: itemsPerPage,
			},
		})

		if err != nil {
			slog.WarnContext(ctx, "failed to fetch github user repositories",
				slog.String("error", err.Error()),
			)
		}

		repos = append(repos, repositoriesToMetadata(r)...)

		slog.DebugContext(ctx, "fetched github user repositories",
			slog.String("username", username),
			slog.String("items_total", fmt.Sprintf("%d (+%d)", len(repos), len(r))),
			slog.Int("page", page),
		)

		if len(r) < itemsPerPage {
			break
		}

		page++
	}

	slog.InfoContext(ctx, "successfully fetched github user repositories",
		slog.String("username", username),
		slog.Int("items_total", len(repos)),
	)

	return nil, nil
}

func (service GithubServiceImpl) GetRepositoryContributors(ctx context.Context, owner string, name string) ([]*entity.Developer, error) {
	if owner == "" || name == "" {
		return nil, fmt.Errorf("repository owner and name are required")
	}

	var page = 1
	var contributors []*entity.Developer

	slog.DebugContext(ctx, "fetching github repository contributors...",
		slog.String("repository_path", fmt.Sprintf("%s/%s", strings.ToLower(owner), strings.ToLower(name))),
	)

	for page <= maxPages {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		slog.DebugContext(ctx, "fetching github repository contributors...",
			slog.String("repository_path", fmt.Sprintf("%s/%s", strings.ToLower(owner), strings.ToLower(name))),
			slog.Int("items_per_page", itemsPerPage),
			slog.Int("page", page),
		)

		contrib, _, err := service.c.Repositories.ListContributors(ctx, strings.ToLower(owner), strings.ToLower(name), &github.ListContributorsOptions{
			Anon: "true",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: itemsPerPage,
			},
		})

		if err != nil {
			slog.WarnContext(ctx, "failed to fetch github repository contributors",
				slog.String("error", err.Error()),
			)

			return nil, err
		}

		contributors = append(contributors, contributorsToEntity(contrib)...)

		slog.DebugContext(ctx, "fetched github repository contributors",
			slog.String("repository_path", fmt.Sprintf("%s/%s", strings.ToLower(owner), strings.ToLower(name))),
			slog.String("items_total", fmt.Sprintf("%d (+%d)", len(contributors), len(contrib))),
			slog.Int("page", page),
		)

		if len(contrib) < itemsPerPage {
			break
		}

		page++
	}

	slog.InfoContext(ctx, "successfully fetched github repository contributors",
		slog.String("repository_path", fmt.Sprintf("%s/%s", strings.ToLower(owner), strings.ToLower(name))),
		slog.Int("items_total", len(contributors)),
	)

	return contributors, nil
}

func (service GithubServiceImpl) GetRepositoryLanguages(ctx context.Context, owner string, name string) (map[string]int, error) {
	if owner == "" || name == "" {
		return nil, fmt.Errorf("repository owner and name are required")
	}

	slog.DebugContext(ctx, "fetching github repository languages...",
		slog.String("repository_path", fmt.Sprintf("%s/%s", strings.ToLower(owner), strings.ToLower(name))),
	)

	langs, _, err := service.c.Repositories.ListLanguages(ctx, strings.ToLower(owner), strings.ToLower(name))
	if err != nil {
		slog.WarnContext(ctx, "failed to fetch github repository languages",
			slog.String("error", err.Error()),
		)

		return nil, err
	}

	slog.InfoContext(ctx, "successfully fetched github repository stargazers",
		slog.String("repository_path", fmt.Sprintf("%s/%s", strings.ToLower(owner), strings.ToLower(name))),
		slog.Int("languages", len(langs)),
	)

	return langs, nil
}

// === Github Repository Issues ===

func (service GithubServiceImpl) GetRepositoryIssues(ctx context.Context, owner, name string) ([]*entity.GithubIssue, error) {
	if owner == "" || name == "" {
		return nil, fmt.Errorf("repository owner and name are required")
	}

	var page = 1
	var issues []*entity.GithubIssue

	slog.DebugContext(ctx, "fetching github repository issues...",
		slog.String("repository_id", fmt.Sprintf("%s/%s", strings.ToLower(owner), strings.ToLower(name))),
	)

	for page <= maxPages {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		iss, _, err := service.c.Issues.ListByRepo(ctx, owner, name, &github.IssueListByRepoOptions{
			Sort:      "created", // recent first
			Direction: "desc",
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: itemsPerPage,
			},
		})

		if err != nil {
			slog.WarnContext(ctx, "failed to fetch github repository issues",
				slog.String("error", err.Error()),
			)

			return nil, err
		}

		issues = append(issues, issuesToEntity(iss)...)

		slog.DebugContext(ctx, "fetched github repository issues",
			slog.String("repository_path", fmt.Sprintf("%s/%s", strings.ToLower(owner), strings.ToLower(name))),
			slog.String("items_total", fmt.Sprintf("%d (+%d)", len(issues), len(iss))),
			slog.Int("page", page),
		)

		if len(issues) < itemsPerPage {
			break
		}

		page++
	}

	return issues, nil
}

// === GitHub Users ===

func (service GithubServiceImpl) GetUserByUsername(ctx context.Context, username string) (*entity.Developer, error) {
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
func (service GithubServiceImpl) GetUserActivity(ctx context.Context, username string) (*entity.Activity, error) {
	if username == "" {
		return nil, fmt.Errorf("github username is required")
	}

	var page = 1
	var events []*github.Event

	for page <= maxPages {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		slog.DebugContext(ctx, "fetching github user activity...",
			slog.String("username", username),
			slog.Int("items_per_page", itemsPerPage),
			slog.Int("page", page),
		)

		evt, _, err := service.c.Activity.ListEventsPerformedByUser(ctx, username, true, &github.ListOptions{
			PerPage: itemsPerPage,
			Page:    page,
		})

		if err != nil {
			slog.WarnContext(ctx, "failed to fetch github user activity",
				slog.String("error", err.Error()),
			)
		}

		events = append(events, evt...)

		slog.DebugContext(ctx, "fetched github user activity",
			slog.String("username", username),
			slog.String("items_total", fmt.Sprintf("%d (+%d)", len(events), len(evt))),
			slog.Int("page", page),
		)

		if len(evt) < itemsPerPage {
			break
		}

		page++
	}

	slog.InfoContext(ctx, "successfully fetched github user activity",
		slog.String("username", username),
		slog.Int("items_total", len(events)),
	)

	return nil, nil
}

// === GitHub Companies ===

func (service GithubServiceImpl) GetCompanyUsers(ctx context.Context, name string) (*entity.Activity, error) {
	if name == "" {
		return nil, fmt.Errorf("github organization company is required")
	}

	var users []*github.User
	var query = fmt.Sprintf("company:%s", name)
	var opts = &github.SearchOptions{
		Sort:  "joined",
		Order: "asc",
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: itemsPerPage,
		},
	}

	for opts.Page <= maxPages {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		slog.DebugContext(ctx, "fetching github company users...",
			slog.String("name", name),
			slog.Int("items_per_page", itemsPerPage),
			slog.Int("page", opts.Page),
		)

		res, _, err := service.c.Search.Users(ctx, query, opts)
		if err != nil {
			slog.WarnContext(ctx, "failed to fetch github company users",
				slog.String("error", err.Error()),
			)
		} else if len(res.Users) == 0 {
			break
		}

		users = append(users, res.Users...)

		slog.DebugContext(ctx, "fetched github company users",
			slog.String("name", name),
			slog.String("items_total", fmt.Sprintf("%d (+%d)", len(users), len(res.Users))),
			slog.Int("page", opts.Page),
		)

		if len(users) >= res.GetTotal() {
			break
		}

		opts.Page++
	}

	slog.InfoContext(ctx, "successfully fetched github company users",
		slog.String("name", name),
		slog.Int("items_total", len(users)),
	)

	return nil, nil
}

// === GitHub Organizations ===

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

func (service GithubServiceImpl) GetOrganizationUsers(ctx context.Context, name string) ([]*entity.Developer, error) {
	if name == "" {
		return nil, fmt.Errorf("github organization name is required")
	}

	var page = 1
	var users []*entity.Developer

	for page <= maxPages {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		slog.DebugContext(ctx, "fetching github organization members...",
			slog.String("name", name),
			slog.Int("items_per_page", itemsPerPage),
			slog.Int("page", page),
		)

		usr, _, err := service.c.Organizations.ListMembers(ctx, name, &github.ListMembersOptions{
			PublicOnly: true,
			Filter:     "all",
			Role:       "all",
			ListOptions: github.ListOptions{
				PerPage: itemsPerPage,
				Page:    page,
			},
		})

		if err != nil {
			slog.WarnContext(ctx, "failed to fetch github organization members",
				slog.String("error", err.Error()),
			)
		}

		users = append(users, usersToEntity(usr)...)

		slog.DebugContext(ctx, "fetched github organization members",
			slog.String("name", name),
			slog.String("items_total", fmt.Sprintf("%d (+%d)", len(users), len(usr))),
			slog.Int("page", page),
		)

		if len(usr) < itemsPerPage {
			break
		}

		page++
	}

	slog.InfoContext(ctx, "successfully fetched github organization members",
		slog.String("name", name),
		slog.Int("items_total", len(users)),
	)

	return users, nil
}
