package git_service

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/storage/memory"
)

type GitHistoryServiceImpl struct {
	config   cfg.GitHistoryInspectionConfig
	basePath string
}

func NewGitHistoryServiceImpl(config cfg.GitHistoryInspectionConfig) *GitHistoryServiceImpl {
	impl := GitHistoryServiceImpl{
		config: config,
	}

	if config.SaveToDisk {
		if config.Filepath == "" {
			slog.Error("git projects directory path is required")
			panic("git projects directory path is required")
		}

		abs, err := filepath.Abs(config.Filepath)
		if err != nil {
			slog.Error("failed to get the absolute path of git projects directory", slog.String("error", err.Error()))
			panic(err)
		}

		impl.basePath = abs
		slog.Info("git projects will be downloaded on disk", slog.String("absolute_path", abs))
	} else {
		slog.Warn("git projects storing on disk is disabled, git history will only be stored in memory and re-downloaded on every scan")
	}

	return &impl
}

func (service *GitHistoryServiceImpl) InspectRepository(ctx context.Context, link *url.URL, branch string) (*entity.Repository, error) {
	if link == nil {
		return nil, errors.New("git project url is required")
	}

	var opts = git.CloneOptions{
		URL:               link.String(),
		Depth:             service.config.MaxDepth,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth, // 10
		Bare:              true,
		NoCheckout:        true,
		ShallowSubmodules: true,
		AllowEmptyRepo:    true,
	}

	if branch != "" {
		opts.ReferenceName = plumbing.NewBranchReferenceName(branch)
		opts.SingleBranch = true
	}

	if service.config.FullClone {
		opts.Bare = false
		opts.NoCheckout = false
	}

	var (
		repo *git.Repository
		err  error
	)

	defer func() {
		if repo == nil {
			return
		}

		err = repo.Close()
		if err != nil {
			slog.WarnContext(ctx, "failed to close git history file",
				slog.String("error", err.Error()),
			)
		}
	}()

	if service.config.SaveToDisk {
		path := filepath.Join(service.basePath, link.Path)
		if !fs.ValidPath(path) {
			return nil, errors.New("cannot create directory, git project path is invalid")
		}

		if checkIfDirectoryExists(path) {
			slog.DebugContext(ctx, "opening existing git repository...",
				slog.String("git_url", link.Redacted()),
				slog.String("git_ref", branch),
			)

			repo, err = git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
		} else {
			slog.DebugContext(ctx, "cloning git repository to disk...",
				slog.String("git_url", link.Redacted()),
				slog.String("git_ref", branch),
				slog.Bool("full_clone", service.config.FullClone),
			)

			repo, err = git.PlainClone(path, &opts)
		}
	} else {
		slog.DebugContext(ctx, "cloning git repository to memory...",
			slog.String("git_url", link.Redacted()),
			slog.String("git_ref", branch),
			slog.Bool("full_clone", service.config.FullClone),
		)

		repo, err = git.Clone(memory.NewStorage(), nil, &opts)
	}

	if err != nil {
		slog.ErrorContext(ctx, "failed to clone git repository",
			slog.String("git_url", link.Redacted()),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("failed to clone git repository: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get git repository head",
			slog.String("git_url", link.Redacted()),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("failed to get git repository head: %w", err)
	}

	slog.DebugContext(ctx, "successfully cloned git repository",
		slog.String("git_url", link.Redacted()),
		slog.Group("data",
			slog.String("head", head.String()),
			slog.String("commit_hash", head.Hash().String()),
		),
	)

	slog.InfoContext(ctx, "inspecting commit history...",
		slog.String("git_url", link.Redacted()),
	)

	err = service.inspectCommitHistory(ctx, repo)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (service *GitHistoryServiceImpl) inspectCommitHistory(ctx context.Context, repo *git.Repository) error {
	commitIter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return err
	}

	var (
		latestCommitAt time.Time
		oldestCommitAt time.Time

		totalCommits uint64
		authors      map[string]uint64
	)

	authors = make(map[string]uint64)

	latestCommit, err := commitIter.Next()
	if err != nil {
		return err
	}

	latestCommitAt = latestCommit.Author.When

	err = commitIter.ForEach(func(c *object.Commit) error {
		v, ok := authors[c.Author.String()]
		if ok {
			authors[c.Author.String()] = v + 1
		} else {
			authors[c.Author.String()] = 1
		}

		oldestCommitAt = c.Author.When
		totalCommits++
		return nil
	})

	slog.DebugContext(ctx, "successfully inspected commit history",
		slog.Uint64("total_commits", totalCommits),
		slog.Time("latest_commit_at", latestCommitAt),
		slog.Time("oldest_commit_at", oldestCommitAt),
	)

	return err
}

func checkIfDirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return true
		}

		return false
	}

	return false
}
