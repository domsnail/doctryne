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

type GithubServiceImpl struct {
	c *github.Client
}

type GithubServiceOpts struct {
	Timeout  time.Duration
	ProxyURL *url.URL

	AccessToken string
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

	_, _, err := service.c.Repositories.Get(ctx, strings.ToLower(owner), strings.ToLower(name))
	if err != nil {
		slog.DebugContext(ctx, "failed to fetch repository",
			slog.String("repository_path", fmt.Sprintf("%s/%s", strings.ToLower(owner), strings.ToLower(name))),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	var repository entity.Repository
	return &repository, nil
}

func (service GithubServiceImpl) GetRepositoryByURL(ctx context.Context, link url.URL) (*entity.Repository, error) {
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

func (service GithubServiceImpl) GetUserByUsername(ctx context.Context, username string) (*entity.Person, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	} else if username == "" {
		return nil, fmt.Errorf("github username is required")
	}

	slog.DebugContext(ctx, "fetching github user...",
		slog.String("username", username),
	)

	_, _, err := service.c.Users.Get(ctx, username)
	if err != nil {
		slog.DebugContext(ctx, "failed to fetch github user",
			slog.String("username", username),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("failed to fetch github user: %w", err)
	}

	var person entity.Person
	return &person, nil
}

func repositoryFromURL(link url.URL) (owner, name string, err error) {
	slog.Debug("trying to determine github repository owner/name...",
		slog.String("repository_url", link.Redacted()),
	)

	switch link.Hostname() {
	case "github.com":
		parts := strings.SplitN(link.Path, "/", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("repository url not supported: '%s'", link.String())
		}

		return parts[0], parts[1], nil
	default:
		return "", "", fmt.Errorf("unsupported git hostname '%s'", link.Hostname())
	}
}
