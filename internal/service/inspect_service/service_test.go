package inspect_service

import (
	"bytes"
	"context"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestFindManifestFiles(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	tmpDir, err := os.MkdirTemp("", "findmanifest_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Build test structure:
	// tmpDir/
	//   package.json                 (depth 0)
	//   go.mod                       (depth 0)
	//   sub1/
	//     package.json               (depth 1)
	//     sub2/
	//       go.mod                   (depth 2)
	//       package.json             (depth 2)
	//   sub3/
	//     empty.txt                  (depth 1, no manifest)

	// Depth 0 files
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("Failed to create package.json at depth 0: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0o644); err != nil {
		t.Fatalf("Failed to create go.mod at depth 0: %v", err)
	}

	// Depth 1
	sub1 := filepath.Join(tmpDir, "sub1")
	if err := os.Mkdir(sub1, 0o755); err != nil {
		t.Fatalf("Failed to create sub1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub1, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("Failed to create package.json at depth 1: %v", err)
	}

	// Depth 2
	sub2 := filepath.Join(sub1, "sub2")
	if err := os.Mkdir(sub2, 0o755); err != nil {
		t.Fatalf("Failed to create sub2: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub2, "go.mod"), []byte("module test2"), 0o644); err != nil {
		t.Fatalf("Failed to create go.mod at depth 2: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sub2, "package.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("Failed to create package.json at depth 2: %v", err)
	}

	// Depth 1 with no manifest
	sub3 := filepath.Join(tmpDir, "sub3")
	if err := os.Mkdir(sub3, 0o755); err != nil {
		t.Fatalf("Failed to create sub3: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sub3, "empty.txt"), []byte("no manifest"), 0o644); err != nil {
		t.Fatalf("Failed to create empty.txt: %v", err)
	}

	config := cfg.NewConfigWithDefaultValues()
	cfg.SetGlobalConfig(config)

	service := InspectionService{}

	synctest.Test(t, func(t *testing.T) {
		manifests, err := service.searchManifestsInDir(context.Background(), bytes.NewReader([]byte(tmpDir)))
		require.NoError(t, err)
		require.Len(t, manifests, 5)
	})

	config.Scan.FileSearchDepth = 2

	synctest.Test(t, func(t *testing.T) {
		manifests, err := service.searchManifestsInDir(context.Background(), bytes.NewReader([]byte(tmpDir)))
		require.NoError(t, err)
		require.Len(t, manifests, 3)
	})

	config.Scan.FileSearchDepth = 1

	synctest.Test(t, func(t *testing.T) {
		manifests, err := service.searchManifestsInDir(context.Background(), bytes.NewReader([]byte(tmpDir)))
		require.NoError(t, err)
		require.Len(t, manifests, 2)
	})
}

func TestRepositoryDedupe(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	t.Run("test dedupe with different repositories", func(t *testing.T) {
		var inspection = entity.Inspection{
			Packages: []*entity.Package{
				{
					Name: "package-1",
					Git: &entity.Git{
						Repository: &entity.Repository{
							Name: "test/repository-1",
							GithubMetadata: &entity.GitHubRepositoryMetadata{
								ID:     1,
								GitURL: nil,
							},
						},
					},
				},
				{
					Name: "package-21",
					Git: &entity.Git{
						Repository: &entity.Repository{
							Name: "test/repository-2",
							GithubMetadata: &entity.GitHubRepositoryMetadata{
								ID:     2,
								GitURL: nil,
							},
						},
					},
				},
				{
					Name: "package-3",
					Git: &entity.Git{
						Repository: &entity.Repository{
							Name: "test/repository-3",
							GithubMetadata: &entity.GitHubRepositoryMetadata{
								ID:     3,
								GitURL: nil,
							},
						},
					},
				},
			},
		}

		dedupeRepositories(context.Background(), &inspection)
		require.Len(t, inspection.Repositories, 3)
	})

	t.Run("test dedupe repositories with same id (no url)", func(t *testing.T) {
		var inspection = entity.Inspection{
			Packages: []*entity.Package{
				{
					Name: "package-1",
					Git: &entity.Git{
						Repository: &entity.Repository{
							Name: "test/repository-1",
							GithubMetadata: &entity.GitHubRepositoryMetadata{
								ID:     1,
								GitURL: nil,
							},
						},
					},
				},
				{
					Name: "package-21",
					Git: &entity.Git{
						Repository: &entity.Repository{
							Name: "test/repository-2",
							GithubMetadata: &entity.GitHubRepositoryMetadata{
								ID:     2,
								GitURL: nil,
							},
						},
					},
				},
				{
					Name: "package-3",
					Git: &entity.Git{
						Repository: &entity.Repository{
							Name: "test/repository-3(2)",
							GithubMetadata: &entity.GitHubRepositoryMetadata{
								ID:     2,
								GitURL: nil,
							},
						},
					},
				},
			},
		}

		dedupeRepositories(context.Background(), &inspection)
		require.Len(t, inspection.Repositories, 3)
	})

	t.Run("test dedupe repositories with same id (same url)", func(t *testing.T) {
		u, _ := url.Parse("https://github.com/test/repository-1-3")

		var inspection = entity.Inspection{
			Packages: []*entity.Package{
				{
					Name: "package-1",
					Git: &entity.Git{
						Repository: &entity.Repository{
							Name: "test/repository-1",
							GithubMetadata: &entity.GitHubRepositoryMetadata{
								ID:     1,
								GitURL: u,
							},
						},
					},
				},
				{
					Name: "package-2",
					Git: &entity.Git{
						Repository: &entity.Repository{
							Name: "test/repository-2",
							GithubMetadata: &entity.GitHubRepositoryMetadata{
								ID:     2,
								GitURL: nil,
							},
						},
					},
				},
				{
					Name: "package-3",
					Git: &entity.Git{
						Repository: &entity.Repository{
							Name: "test/repository-3(1)",
							GithubMetadata: &entity.GitHubRepositoryMetadata{
								ID:     2,
								GitURL: u,
							},
						},
					},
				},
			},
		}

		dedupeRepositories(context.Background(), &inspection)
		require.Len(t, inspection.Repositories, 2)
		require.Same(t, inspection.Packages[0].Git.Repository, inspection.Packages[2].Git.Repository, "must point to same repository")
	})

}
