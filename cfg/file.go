package cfg

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func NewConfigFromFile(path string) (*Config, error) {
	var file *os.File
	defer func() {
		_ = file.Close()
	}()

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute config file path: %w", err)
	}

	file, err = os.Open(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	config.FilePath = absolutePath

	switch filepath.Ext(absolutePath) {
	case ".yml", ".yaml":
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal yaml/yml config file: %w", err)
		}
	case ".json":
		err = json.Unmarshal(data, &config)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal json config file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file extension: %s", filepath.Ext(absolutePath))
	}

	return nil, nil
}
