package cfg

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/domsnail/doctryne/pkg/types"
)

func NewConfigFromFlags(ctx context.Context) (config *Config, err error) {
	config = NewConfigWithDefaultValues()

	// === Configuration file
	configFile := flag.String("config", "", "configuration file path")

	// === Output
	format := flag.String("format", "json", "output report format")

	// === Logging setup
	logLevel := flag.Int("log-level", 0, "log level (-4, 8)")
	logFormat := flag.String("log-format", "text", "log format (text, json)")

	// === Server setup
	server := flag.Bool("server", false, "use server mode")
	host := flag.String("host", "0.0.0.0", "server address")
	port := flag.Uint64("port", 80, "server port")
	accessKey := flag.String("access-key", "", "server access key")

	// === Client setup
	proxy := flag.String("http-proxy", "", "http proxy server")
	insecure := flag.Bool("insecure", false, "use insecure http")
	timeout := flag.Duration("timeout", time.Second*30, "operation timeout")

	flag.Parse()

	if configFile != nil && *configFile != "" {
		slog.InfoContext(ctx, "loading configuration from file, cli flags will be ignored", slog.String("file_path", *configFile))
		return NewConfigFromFile(*configFile)
	}

	if format != nil && *format != "" {
		config.Output.Format = types.ReportFormat(*format)
	}

	if server != nil && *server {
		config.Server.Enabled = true

		config.Server.Host = *host
		config.Server.Port = uint32(*port)

		config.Server.AccessKey = *accessKey
	}

	if insecure != nil {
		config.Insecure = *insecure
	}

	if proxy != nil && *proxy != "" {
		config.HttpProxy, err = url.Parse(*proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy url: %w", err)
		}
	}

	if logFormat != nil && *logFormat != "" {
		config.Logging.Format = *logFormat
	}

	if logLevel != nil && *logLevel != 0 {
		config.Logging.Level = *logLevel
	}

	if timeout != nil && *timeout > 0 {
		config.Timeout = *timeout
	}

	err = config.IsValid()
	if err != nil {
		return nil, errors.New("invalid configuration: " + err.Error())
	}

	return config, nil
}
