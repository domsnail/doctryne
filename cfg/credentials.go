package cfg

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

const secretsDir = "/run/secrets"

type CredentialsConfig struct {
	GithubApiKey        string `json:"github_api_key" yaml:"github_api_key"`
	NpmApiKey           string `json:"npm_api_key" yaml:"npm_api_key"`
	StackExchangeApiKey string `json:"stack_exchange_api_key" yaml:"stack_exchange_api_key"`

	NVDApiKey string `json:"nvd_api_key" yaml:"nvd_api_key"`
}

func LoadCredentialsFromSecrets() (*CredentialsConfig, error) {
	slog.Info("loading credentials from docker secrets...")

	entries, err := os.ReadDir(secretsDir)
	if errors.Is(err, os.ErrNotExist) {
		slog.Warn("credentials will not be loaded from docker secrets, directory does not exist")
		return nil, nil
	}

	slog.Info("found multiple secret files in directory",
		slog.Int("secrets_dir", len(entries)),
		slog.Int("total_files", len(entries)),
	)

	var c CredentialsConfig

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		var buf []byte
		var err error

		path := filepath.Join(secretsDir, entry.Name())
		buf, err = os.ReadFile(path)
		if err != nil {
			slog.Warn(fmt.Sprintf("failed to read secret file '%s': %s, skipping file...", path, err.Error()),
				slog.String("full_path", path),
			)

			continue
		} else if len(buf) == 0 {
			slog.Warn(fmt.Sprintf("secret file '%s' is empty, skipping file...", path),
				slog.String("full_path", path),
			)

			continue
		}

		switch filepath.Base(entry.Name()) {
		case "github-api-key":
			c.GithubApiKey = string(buf)
		case "npm-api-key":
			c.NpmApiKey = string(buf)
		case "stack-exchange-api-key":
			c.StackExchangeApiKey = string(buf)
		case "nvd-api-key":
			c.NVDApiKey = string(buf)
		default:
			slog.Warn("unknown secret type",
				slog.String("secret_name", entry.Name()),
			)

			continue
		}

		slog.Info("successfully loaded key from secret file",
			slog.String("secret_name", entry.Name()),
		)
	}

	return &c, nil
}
