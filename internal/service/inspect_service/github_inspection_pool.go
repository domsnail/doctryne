package inspect_service

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
	"github.com/domsnail/doctryne/internal/types"
)

type GitHubInspectionPool struct {
	github service.IGithubService
	opts   GitHubInspectionOptions

	wg  *sync.WaitGroup
	ctx context.Context

	capacity int32
	active   atomic.Int32

	c *sync.Cond
}

func NewGitHubInspectionPool(ctx context.Context, github service.IGithubService, opts GitHubInspectionOptions) *GitHubInspectionPool {
	var mu sync.Mutex

	wg := &sync.WaitGroup{}

	return &GitHubInspectionPool{
		github:   github,
		opts:     opts,
		capacity: cfg.GlobalConfig.Concurrency,
		wg:       wg,
		ctx:      ctx,
		c:        sync.NewCond(&mu),
	}
}

// GitHubInspectionOptions overrides application default settings from cfg.ScanConfig
type GitHubInspectionOptions struct {
	Mode types.InspectionMode

	DeepRepositoryInspection   bool
	ExtractFullContributorInfo bool
	InspectIssues              bool
}

func (pool *GitHubInspectionPool) Inspect(pkg *entity.Package) {
	pool.c.L.Lock()
	for pool.active.Load() >= pool.capacity {
		pool.c.Wait()
	}

	pool.active.Add(1)
	pool.c.L.Unlock()

	pool.wg.Go(func() {
		defer func() {
			pool.c.L.Lock()
			pool.active.Add(-1)
			pool.c.L.Unlock()

			pool.c.Signal()
		}()

		gitUrl := pkg.GetGitURL()
		if gitUrl == nil {
			slog.ErrorContext(pool.ctx, "failed to github package page",
				slog.String("error", "package git url is nil"),
				slog.String("details", "git url must be defined in registry metadata, this is not intended behavior"),
			)

			return
		}

		slog.DebugContext(pool.ctx, "inspecting github package page...",
			slog.String("package_name", pkg.Name),
			slog.String("package_version", pkg.Version),
			slog.String("git_url", gitUrl.Redacted()),
		)

		repo, err := pool.github.GetRepositoryByURL(pool.ctx, gitUrl)
		if err != nil {
			slog.WarnContext(pool.ctx, "failed to inspect github package page",
				slog.String("package_name", pkg.Name),
				slog.String("package_version", pkg.Version),
				slog.String("git_url", gitUrl.Redacted()),
				slog.String("error", err.Error()),
			)

			return
		}

		var gitMetadata = entity.Git{
			Url: gitUrl,
			Repository: &entity.Repository{
				Name:           repo.Name,
				GitURL:         gitUrl,
				GithubID:       repo.ID,
				GithubMetadata: repo,
				CreatedAt:      repo.CreatedAt,
				UpdatedAt:      repo.UpdatedAt,
				PushedAt:       repo.PushedAt,
			},
		}

		if repo.Org != nil {
			// fetching full organization info
			repo.Org, err = pool.github.GetOrganizationByName(pool.ctx, repo.Org.GithubMetadata.Login)
			if err != nil {
				slog.WarnContext(pool.ctx, "failed to inspect github organization page",
					slog.String("organization_name", repo.Org.Name),
					slog.String("error", err.Error()),
				)

				return
			}
		}

		repo.Languages, err = pool.github.GetRepositoryLanguages(pool.ctx, repo.Owner.Username, repo.Name)
		if err != nil {
			slog.WarnContext(pool.ctx, "failed to fetch github repository languages",
				slog.String("repository_name", repo.Owner.Username+"/"+repo.Name),
				slog.String("error", err.Error()),
			)
		}

		repo.Contributors, err = pool.github.GetRepositoryContributors(pool.ctx, repo.Owner.Username, repo.Name)
		if err != nil {
			slog.WarnContext(pool.ctx, "failed to fetch github repository contributors",
				slog.String("repository_name", repo.Owner.Username+"/"+repo.Name),
				slog.String("error", err.Error()),
			)
		}

		if pool.opts.InspectIssues {
			repo.Issues, err = pool.github.GetRepositoryIssues(pool.ctx, repo.Owner.Username, repo.Name)
			if err != nil {
				slog.WarnContext(pool.ctx, "failed to fetch github repository issues",
					slog.String("repository_name", repo.Owner.Username+"/"+repo.Name),
					slog.String("error", err.Error()),
				)
			}
		}

		pkg.Git = &gitMetadata
	})
}

func (pool *GitHubInspectionPool) inspectRepositoryContributors(contributors []*entity.GithubDeveloperProfile) error {
	return nil
}

func (pool *GitHubInspectionPool) Wait() error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		pool.wg.Wait()
	}()

	select {
	case <-pool.ctx.Done():
		return pool.ctx.Err()
	case <-done:
		return nil
	}
}
