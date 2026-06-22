package github_service

import (
	"context"
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
const maxPages = 3

type GithubServiceImpl struct {
	c *github.Client
}

type GithubServiceOpts struct {
	Timeout time.Duration

	AccessToken string

	LatestActivityPeriod time.Duration
}

const defaultLatestActivityPeriod = 30 * 24 * time.Hour

func NewGithubServiceImpl(opts GithubServiceOpts) *GithubServiceImpl {
	slog.Debug("initializing github client...",
		slog.Bool("using_access_token", opts.AccessToken != ""),
	)

	if opts.Timeout <= 0 {
		slog.Debug("timeout is not set or invalid, setting default timeout for github client...",
			slog.Duration("timeout", http.DefaultClient.Timeout),
		)

		opts.Timeout = http.DefaultClient.Timeout
	}

	if opts.LatestActivityPeriod <= 0 {
		slog.Debug("latest activity is not set or invalid, setting default period...",
			slog.Duration("period", defaultLatestActivityPeriod),
		)

		opts.LatestActivityPeriod = defaultLatestActivityPeriod
	}

	var (
		client *github.Client
		err    error
	)

	var githubClientOpts = []github.ClientOptionsFunc{
		github.WithTimeout(opts.Timeout),
		github.WithTransport(http.DefaultTransport),
	}

	if opts.AccessToken != "" {
		githubClientOpts = append(githubClientOpts, github.WithAuthToken(opts.AccessToken))
	}

	client, err = github.NewClient(githubClientOpts...)
	if err != nil {
		slog.Error("failed to initialize github client", slog.String("error", err.Error()))
	}

	return &GithubServiceImpl{c: client}
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

func (service GithubServiceImpl) GetUserOwnedRepositories(ctx context.Context, username string) ([]*entity.Repository, error) {
	if username == "" {
		return nil, fmt.Errorf("github username is required")
	}

	var page = 1
	var repos []*entity.Repository

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

		repos = append(repos, repositoriesToEntity(r)...)

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

func (service GithubServiceImpl) GetRepositoryContributors(ctx context.Context, owner, name string) ([]*entity.Developer, error) {
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
