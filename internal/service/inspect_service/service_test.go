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

func TestDeveloperDedupe(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	var inspection = entity.Inspection{
		Packages: []*entity.Package{
			{
				Name: "test1",
				RegistryMetadata: &entity.RegistryMetadata{
					Contributors: &entity.AffiliatedDevelopers{
						Authors: []*entity.Developer{
							{
								Name:     "test 1",
								Username: "test-1",
								Emails:   []string{"test1@test.com"},
							},
						},
						Contributors: []*entity.Developer{
							{
								Name:     "test 2",
								Username: "test-2",
							},
						},
					},
				},
			},
			{
				Name: "test2",
				RegistryMetadata: &entity.RegistryMetadata{
					Contributors: &entity.AffiliatedDevelopers{
						Authors: []*entity.Developer{
							{
								Name:     "test 3",
								Username: "test-3",
							},
						},
						CodeOwners: []*entity.Developer{
							{
								Name:     "test 1",
								Username: "test-1",
								Emails:   []string{"test1@test.com", "test11@test.com"},
							},
							{
								Name:     "test 4",
								Username: "test-4",
							},
						},
					},
				},
			},
			{
				Name: "test3",
				RegistryMetadata: &entity.RegistryMetadata{
					Contributors: &entity.AffiliatedDevelopers{
						Authors: []*entity.Developer{
							{
								Name:     "test 2",
								Username: "test-2",
							},
						},
						CodeOwners: []*entity.Developer{
							{
								Name:     "test 2",
								Username: "test-2",
								GithubID: 2,
								GithubMetadata: &entity.GithubDeveloperProfile{
									ID:     2,
									NodeID: "test-node-id-2",
								},
							},
							{
								Name:     "test 5",
								Username: "test-5",
								GithubID: 5,
								GithubMetadata: &entity.GithubDeveloperProfile{
									ID:     5,
									NodeID: "test-node-id-5",
								},
							},
						},
					},
				},
			},
			{
				Name: "test4",
				RegistryMetadata: &entity.RegistryMetadata{
					Contributors: &entity.AffiliatedDevelopers{
						Contributors: []*entity.Developer{
							{
								Name:     "test 5",
								Username: "test-5",
							},
							{
								Name:     "test 2",
								Username: "test-2",
								GithubID: 22,
								GithubMetadata: &entity.GithubDeveloperProfile{
									ID: 22,
								},
							},
						},
						CodeOwners: []*entity.Developer{
							{
								Name:     "test 1",
								Username: "test-1",
							},
							{
								Name:     "test 1",
								Username: "test-1",
								Emails:   []string{"test111@test.com"},
							},
						},
					},
				},
			},
		},
		Repositories: []*entity.Repository{
			{
				Name: "test1",
				Commiters: []*entity.Developer{
					{
						Name:     "test 6",
						Username: "test-6",
					},
					{
						Name:     "test 7",
						Username: "test-7",
					},
					{
						Name:     "test 1",
						Username: "test-1",
					},
				},
			},
			{
				Name: "test2",
				GithubMetadata: &entity.GitHubRepositoryMetadata{
					ID: 2,
					Owner: &entity.GithubDeveloperProfile{
						Fullname: "test 5",
						Username: "test-5",
						ID:       5,
						Emails:   []string{"test5@test.com"},
					},
					Org: nil,
					Contributors: []*entity.Developer{
						{
							Name:     "test 4",
							Username: "test-4",
						},
						{
							Name:     "test 10",
							Username: "test-10",
						},
					},
				},
				Commiters: []*entity.Developer{
					{
						Name:     "test 8",
						Username: "test-8",
					},
				},
			},
			{
				Name: "test3",
				GithubMetadata: &entity.GitHubRepositoryMetadata{
					ID: 3,
					Owner: &entity.Developer{
						Name:     "test 11",
						Username: "test-11",
					},
					Org: &entity.Organization{
						Name:  "test 1",
						Login: "test-1",
					},
					Contributors: nil,
				},
				Commiters: []*entity.Developer{
					{
						Name:     "test 9",
						Username: "test-9",
					},
				},
			},
			{
				Name: "test4",
				GithubMetadata: &entity.GitHubRepositoryMetadata{
					ID: 4,
					Org: &entity.Organization{
						Name:   "test 2",
						Login:  "test-2",
						Emails: []string{"test2@test.com"},
					},
				},
			},
			{
				Name: "test5",
				GithubMetadata: &entity.GitHubRepositoryMetadata{
					ID: 5,
					Org: &entity.Organization{
						Name:     "test 2",
						Login:    "test-2",
						Emails:   []string{"test2@test.com", "test22@test.com"},
						GithubID: 2,
						GithubMetadata: &entity.GithubOrganizationMetadata{
							ID:   2,
							Name: "test-2",
						},
					},
				},
			},
		},
	}

	uniqueDevelopers, uniqueOrgs := extractAndDedupeAllDevelopers(context.Background(), &inspection)
	require.Len(t, inspection.Packages, 4)
	require.Len(t, inspection.Repositories, 5)
	require.Len(t, uniqueDevelopers, 12, "11 unique developers + 1 conflicting")
	require.Len(t, uniqueOrgs, 2)

	test2 := inspection.Packages[0].RegistryMetadata.Contributors.Contributors[0]
	require.Len(t, inspection.Packages[0].RegistryMetadata.Contributors.Contributors, 1)
	require.NotNil(t, test2.GithubMetadata, "must be filled")
	require.EqualValues(t, 2, test2.GithubID)

	test2_next := inspection.Packages[2].RegistryMetadata.Contributors.CodeOwners[0]
	require.Len(t, inspection.Packages[2].RegistryMetadata.Contributors.CodeOwners, 2)
	require.NotNil(t, test2_next.GithubMetadata, "must be filled")
	require.EqualValues(t, 2, test2_next.GithubID)
	require.Same(t, test2, test2_next)
	require.Same(t, test2, test2_next, inspection.Packages[2].RegistryMetadata.Contributors.Authors[0])

	test5 := inspection.Packages[2].RegistryMetadata.Contributors.CodeOwners[1]
	require.Len(t, inspection.Packages[2].RegistryMetadata.Contributors.CodeOwners, 2)
	require.NotNil(t, test5.GithubMetadata, "must be filled")
	require.EqualValues(t, 5, test5.GithubID)
	require.Len(t, test5.Emails, 1)
	require.Same(t, test5, inspection.Repositories[1].GithubMetadata.Owner)

	test1 := inspection.Packages[0].RegistryMetadata.Contributors.Authors[0]
	require.Len(t, test1.Emails, 3)

	require.Same(t, test1, inspection.Repositories[0].Commiters[2])

	require.Same(t, inspection.Repositories[4].GithubMetadata.Org, inspection.Repositories[3].GithubMetadata.Org)
	require.EqualValues(t, "test-2", inspection.Repositories[3].GithubMetadata.Org.GithubMetadata.Name)
	require.Len(t, inspection.Repositories[3].GithubMetadata.Org.Emails, 2)
	require.EqualValues(t, len(inspection.Repositories[3].GithubMetadata.Org.Emails), len(inspection.Repositories[4].GithubMetadata.Org.Emails))
}
