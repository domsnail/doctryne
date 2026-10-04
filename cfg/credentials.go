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

		buf, err = os.ReadFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read secret file '%s': %w", entry.Name(), err)
		} else if len(buf) == 0 {
			return nil, fmt.Errorf("secret file '%s' is empty", entry.Name())
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
			slog.Warn("unknown secret type", slog.String("secret_name", entry.Name()))
			continue
		}
	}

	return &c, nil
}
