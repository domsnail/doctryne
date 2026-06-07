package manifest_service

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"testing/synctest"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/pkg/types"
	"github.com/stretchr/testify/require"
)

const filesPath = "test_data"

type testCase struct {
	name string

	file     string
	lockfile string

	expected expectedValues
}

type expectedValues struct {
	packageName    string
	packageVersion string

	language types.Language

	discoveredPackages int

	totalDependencies    int
	devDependencies      int
	optionalDependencies int
}

func TestManifestServiceImpl_ProcessManifest(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	})))

	service := ManifestServiceImpl{}

	synctest.Test(t, func(t *testing.T) {
		config := cfg.NewConfigWithDefaultValues()
		config.Languages.JavaScript.CheckDevDependencies = true
		config.Languages.JavaScript.CheckOptionalDependencies = true
		cfg.SetGlobalConfig(config)

		tests := []testCase{
			{
				name: "test cyclonedx ui package.json",
				file: "cyclonedx_ui/package.json",
				expected: expectedValues{
					packageName:          "frontend",
					packageVersion:       "0.0.0",
					language:             types.Language_JavaScript,
					discoveredPackages:   1,
					totalDependencies:    45,
					devDependencies:      13,
					optionalDependencies: 0,
				},
			},
			{
				name:     "test cyclonedx ui package.json with lockfile",
				file:     "cyclonedx_ui/package.json",
				lockfile: "cyclonedx_ui/package-lock.json",
				expected: expectedValues{
					packageName:          "frontend",
					packageVersion:       "0.0.0",
					language:             types.Language_JavaScript,
					discoveredPackages:   1,
					totalDependencies:    375,
					devDependencies:      125,
					optionalDependencies: 78,
				},
			},
		}

		for _, tt := range tests {
			file, err := os.Open(filepath.Join(filesPath, tt.file))
			require.NoError(t, err)
			require.NotNil(t, file)

			manifest := entity.NewManifest().WithFilename(tt.file)
			err = manifest.SetFileContent(file)
			require.NoError(t, err)

			if tt.lockfile != "" {
				lockfile, err := os.Open(filepath.Join(filesPath, tt.lockfile))
				require.NoError(t, err)

				err = manifest.SetLockfileContent(lockfile)
				require.NoError(t, err)
			}

			err = service.ProcessManifest(context.Background(), manifest)
			require.NoError(t, err)
			require.NotNil(t, manifest)
			require.NotNil(t, manifest.UUID)

			require.Equal(t, tt.expected.language, manifest.Language)

			require.EqualValues(t, tt.file, manifest.Metadata.Filename)
			require.NotEmpty(t, manifest.Metadata.Checksum)

			require.EqualValues(t, tt.expected.discoveredPackages, len(manifest.DiscoveredPackages))
			require.EqualValues(t, tt.expected.totalDependencies, len(manifest.DiscoveredPackages[0].Dependencies))

			require.EqualValues(t, tt.expected.devDependencies, manifest.CountDevDependencies())
			require.EqualValues(t, tt.expected.optionalDependencies, manifest.CountOptionalDependencies())
		}
	})

	synctest.Test(t, func(t *testing.T) {
		config := cfg.NewConfigWithDefaultValues()
		config.Languages.JavaScript.CheckDevDependencies = false
		config.Languages.JavaScript.CheckOptionalDependencies = false
		cfg.SetGlobalConfig(config)

		tests := []testCase{
			{
				name: "test cyclonedx ui package.json",
				file: "cyclonedx_ui/package.json",
				expected: expectedValues{
					packageName:          "frontend",
					packageVersion:       "0.0.0",
					language:             types.Language_JavaScript,
					discoveredPackages:   1,
					totalDependencies:    32,
					devDependencies:      0,
					optionalDependencies: 0,
				},
			},
			{
				name:     "test cyclonedx ui package.json with lockfile",
				file:     "cyclonedx_ui/package.json",
				lockfile: "cyclonedx_ui/package-lock.json",
				expected: expectedValues{
					packageName:          "frontend",
					packageVersion:       "0.0.0",
					language:             types.Language_JavaScript,
					discoveredPackages:   1,
					totalDependencies:    182,
					devDependencies:      0,
					optionalDependencies: 0,
				},
			},
		}

		for _, tt := range tests {
			file, err := os.Open(filepath.Join(filesPath, tt.file))
			require.NoError(t, err)
			require.NotNil(t, file)

			manifest := entity.NewManifest().WithFilename(tt.file)
			err = manifest.SetFileContent(file)
			require.NoError(t, err)

			if tt.lockfile != "" {
				lockfile, err := os.Open(filepath.Join(filesPath, tt.lockfile))
				require.NoError(t, err)

				err = manifest.SetLockfileContent(lockfile)
				require.NoError(t, err)
			}

			err = service.ProcessManifest(context.Background(), manifest)
			require.NoError(t, err)
			require.NotNil(t, manifest)
			require.NotNil(t, manifest.UUID)

			require.Equal(t, tt.expected.language, manifest.Language)

			require.EqualValues(t, tt.file, manifest.Metadata.Filename)
			require.NotEmpty(t, manifest.Metadata.Checksum)

			require.EqualValues(t, tt.expected.discoveredPackages, len(manifest.DiscoveredPackages))
			require.EqualValues(t, tt.expected.totalDependencies, len(manifest.DiscoveredPackages[0].Dependencies))
			require.Zero(t, manifest.CountDevDependencies())
			require.Zero(t, manifest.CountOptionalDependencies())
		}
	})
}
