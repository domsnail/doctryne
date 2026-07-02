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

const (
	defaultRemoteName = "origin"
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
		slog.Warn("git projects will be downloaded on disk", slog.String("absolute_path", abs))
	} else {
		slog.Warn("git projects storing on disk is disabled, git history will only be stored in memory and re-downloaded on every scan")
	}

	return &impl
}

func (service *GitHistoryServiceImpl) InspectRepository(ctx context.Context, link *url.URL, branch string) (*entity.Repository, error) {
	if link == nil {
		return nil, errors.New("git project url is required")
	}

	var repo = entity.Repository{}
	var opts = git.CloneOptions{
		RemoteName:        defaultRemoteName,
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
		gitRepo *git.Repository
		err     error
	)

	defer func() {
		if gitRepo == nil {
			return
		}

		err = gitRepo.Close()
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

			// DetectDotGit option must be disabled, or it will find .git in parent dirs
			gitRepo, err = git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: false})
			if err != nil {
				slog.ErrorContext(ctx, "failed to open git repository",
					slog.String("git_url", link.Redacted()),
					slog.String("file_path", path),
					slog.String("error", err.Error()),
				)

				return nil, err
			}

			if service.config.AlwaysFetch { // todo: add fetch age check
				slog.DebugContext(ctx, "fetching git repository...",
					slog.String("git_url", link.Redacted()),
					slog.String("file_path", path),
				)

				err = gitRepo.FetchContext(ctx, &git.FetchOptions{
					RemoteName: defaultRemoteName,
					RemoteURL:  link.String(),
					Depth:      service.config.MaxDepth,
					Force:      true,
					Prune:      true,
				})
			}
		} else {
			slog.DebugContext(ctx, "cloning git repository to disk...",
				slog.String("git_url", link.Redacted()),
				slog.String("git_ref", branch),
				slog.Bool("full_clone", service.config.FullClone),
			)

			gitRepo, err = git.PlainClone(path, &opts)
		}
	} else {
		slog.DebugContext(ctx, "cloning git repository to memory...",
			slog.String("git_url", link.Redacted()),
			slog.String("git_ref", branch),
			slog.Bool("full_clone", service.config.FullClone),
		)

		gitRepo, err = git.Clone(memory.NewStorage(), nil, &opts)
	}

	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		slog.DebugContext(ctx, "git repository already up-to-date")
	} else if err != nil {
		slog.ErrorContext(ctx, "failed to clone git repository",
			slog.String("git_url", link.Redacted()),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("failed to clone git repository: %w", err)
	}

	head, err := gitRepo.Head()
	if err != nil {
		slog.ErrorContext(ctx, "failed to get git repository head",
			slog.String("git_url", link.Redacted()),
			slog.String("error", err.Error()),
		)

		return nil, fmt.Errorf("failed to get git repository head: %w", err)
	}

	slog.DebugContext(ctx, "successfully opened cloned git repository",
		slog.String("git_url", link.Redacted()),
		slog.Group("data",
			slog.String("head", head.String()),
		),
	)

	slog.InfoContext(ctx, "inspecting commit history...",
		slog.String("git_url", link.Redacted()),
	)

	err = service.inspectCommitHistory(ctx, &repo, gitRepo)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// inspectCommitHistory iterates over every git history commit,
// collects every commit author email and name,
// collects statistics of lines additions/deletions for every author
func (service *GitHistoryServiceImpl) inspectCommitHistory(ctx context.Context, r *entity.Repository, g *git.Repository) error {
	commitIter, err := g.Log(&git.LogOptions{})
	if err != nil {
		return err
	}

	var (
		latestCommitAt time.Time
		oldestCommitAt time.Time

		totalStats entity.CommitStats
		commits    []*entity.Commit
	)

	authors := newAuthorsStore()

	latestCommit, err := commitIter.Next()
	if err != nil {
		return err
	}

	latestCommitAt = latestCommit.Author.When

	err = commitIter.ForEach(func(c *object.Commit) error {
		developer := authors.Update(c.Author.Email, c.Author.Name)

		// _ = authors.Update(c.Committer.Email, c.Committer.Name)

		commit := entity.Commit{
			Hash:      c.Hash.String(),
			Message:   c.Message,
			Author:    developer,
			CreatedAt: c.Author.When,
			Stats:     new(entity.CommitStats),
		}

		defer func() {
			commits = append(commits, &commit)

			oldestCommitAt = c.Author.When
		}()

		t := c.Type()
		stats, err := c.Stats()
		if err != nil {
			slog.WarnContext(ctx, "failed to stat commit",
				slog.String("commit_type", t.String()),
				slog.String("commit_hash", c.Hash.String()),
				slog.String("error", err.Error()),
			)

			return nil
		}

		commit.Stats.LinesAdded, commit.Stats.LinesDeleted, commit.Stats.ChangedFiles = processStats(stats)
		authors.AddStats(developer, commit.Stats)
		totalStats.Add(commit.Stats)
		return nil
	})

	slog.DebugContext(ctx, "successfully inspected commit history",
		slog.Int("total_commits", len(commits)),
		slog.Int("total_authors", len(authors.authors)),
		slog.Time("latest_commit_at", latestCommitAt),
		slog.Time("oldest_commit_at", oldestCommitAt),
		slog.Group("total_stats",
			slog.Int("files_changed", len(totalStats.ChangedFiles)),
			slog.Int("lines_added", totalStats.LinesAdded),
			slog.Int("lines_deleted", totalStats.LinesDeleted),
		),
	)

	r.Commits = commits
	r.CommitStats = &totalStats
	r.DeveloperCommitStats = authors.Stats()

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

func processStats(stats object.FileStats) (linesAdded, linesDeleted int, filesChanged []string) {
	for _, s := range stats {
		filesChanged = append(filesChanged, s.Name)

		linesAdded += s.Addition
		linesDeleted += s.Deletion
	}

	return
}
