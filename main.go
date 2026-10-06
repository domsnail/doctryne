package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/cmd"
	"github.com/domsnail/doctryne/internal/repos"
	"github.com/domsnail/doctryne/internal/repos/orm"
	"github.com/domsnail/doctryne/internal/service/developer_service"
	"github.com/domsnail/doctryne/internal/service/github_service"
	"github.com/domsnail/doctryne/internal/service/inspect_service"
	"github.com/domsnail/doctryne/internal/service/manifest_service"
	"github.com/domsnail/doctryne/internal/service/registry_service"
	"github.com/domsnail/doctryne/internal/service/vulnerability_service"
	"github.com/domsnail/doctryne/pkg/alerts"
	"github.com/domsnail/doctryne/pkg/stack_exchange"
	"gorm.io/gorm"
)

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config, err := cfg.NewConfigFromFlags(rootCtx)
	if err != nil {
		panic(err.Error())
	}

	var handler slog.Handler
	switch config.Logging.Format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: config.Logging.AddSource,
			Level:     slog.Level(config.Logging.Level),
		})
	case "text":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: config.Logging.AddSource,
			Level:     slog.Level(config.Logging.Level),
		})
	default:
		panic(fmt.Sprintf("invalid logging format: '%s'", config.Logging.Format))
	}

	if config.Logging.Alerts.Enabled {
		alertConfig := config.Logging.Alerts
		var alertsHandler slog.Handler

		slog.Info("enabling alerting service...",
			slog.String("system", string(alertConfig.System)),
			slog.String("environment", string(alertConfig.Environment)),
		)

		switch alertConfig.System {
		case cfg.AlertsSystem_GitLab:
			hostname, _ := os.Hostname()

			alertsHandler = alerts.NewGitlabAlertClient(
				alerts.WithHTTPClient(http.DefaultClient),
				alerts.WithHost(hostname),
				alerts.WithEndpoint(alertConfig.Endpoint),
				alerts.WithEnvironment(alertConfig.Environment),
				alerts.WithToken(alertConfig.Key),
				alerts.WithMinLevel(slog.Level(alertConfig.MinLevel)),
			)

		default:
			panic(fmt.Sprintf("invalid alerting system provided: %s", alertConfig.System))
		}

		slog.SetDefault(slog.New(slog.NewMultiHandler(handler, alertsHandler)))
		slog.Log(rootCtx, slog.Level(alertConfig.MinLevel), "server has been restarted")
	} else {
		slog.SetDefault(slog.New(handler))
	}

	slog.DebugContext(rootCtx, "loaded configuration variables",
		slog.String("config_file_path", config.FilePath),
		slog.Bool("use_http_proxy", config.HttpProxy != ""),
		slog.Duration("timeout", config.Timeout),
		slog.Group("output",
			slog.String("format", string(config.Output.Format)),
		),
		slog.Group("server",
			slog.Bool("enabled", config.Server.Enabled),
			slog.String("host", config.Server.Host),
			slog.Int("port", int(config.Server.Port)),
			slog.Bool("require_access_key", config.Server.AccessKey != ""),
		),
		slog.Group("logging",
			slog.Int("level", config.Logging.Level),
			slog.String("format", config.Logging.Format),
			slog.Bool("add_source", config.Logging.AddSource),
		),
		slog.Group("database",
			slog.String("driver", config.Database.Driver),
			slog.String("address", net.JoinHostPort(config.Database.Host, fmt.Sprint(config.Database.Port))),
			slog.String("timezone", config.Database.Timezone),
		),
		slog.Group("credentials",
			slog.Bool("has_nvd_access_key", len(config.Credentials.NVDApiKey) > 0),
			slog.Bool("has_github_access_key", len(config.Credentials.GithubApiKey) > 0),
			slog.Bool("has_npm_bearer_token", len(config.Credentials.NpmApiKey) > 0),
			slog.Bool("has_stack_exchange_access_key", len(config.Credentials.StackExchangeApiKey) > 0),
		),
	)

	slog.DebugContext(rootCtx, "setting global config...")
	cfg.SetGlobalConfig(config)

	var (
		conn *gorm.DB
	)

	if !config.HasDatabase() {
		slog.WarnContext(rootCtx, "database not configured, using cached sqlite")
		conn, err = orm.NewDatabaseConn(rootCtx, &cfg.DatabaseConfig{
			Driver: "sqlite",
			File:   "file::memory:?cache=shared",
			Name:   "doctryne",
		})
	} else {
		switch config.Database.Driver {
		case "sqlite3", "sqlite":
			if config.Database.File == "" {
				panic(fmt.Sprintf("sqlite3 database file not set"))
			}

			conn, err = orm.NewDatabaseConn(rootCtx, &config.Database)
		case "postgres", "mysql":
			conn, err = orm.NewDatabaseConn(rootCtx, &config.Database)
		case "local", "cache":
			slog.WarnContext(rootCtx, "database not configured, using cached sqlite")
			conn, err = orm.NewDatabaseConn(rootCtx, &cfg.DatabaseConfig{
				Driver: "sqlite",
				File:   "file::memory:?cache=shared",
				Name:   "doctryne",
			})
		default:
			panic(fmt.Sprintf("unsupported database driver: %s", config.Database.Driver))
		}
	}

	if err != nil {
		panic(fmt.Sprintf("failed to initialize database connection: %s", err.Error()))
	}

	slog.DebugContext(rootCtx, "running migrations...")
	err = orm.AutoMigrate(conn)
	if err != nil {
		panic(fmt.Sprintf("failed to complete database migrations: %s", err.Error()))
	}

	slog.InfoContext(rootCtx, "database migrations completed successfully")

	slog.DebugContext(rootCtx, "initializing services...")
	databaseUpdater, err := vulnerability_service.NewVulnerabilityDatabaseUpdater(conn, config.VulnerabilityDatabase)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize vulnerability database updater: %s", err.Error()))
	}

	vulnerabilityService := vulnerability_service.NewVulnerabilityServiceImpl(
		repos.NewVulnerabilityRepoImpl(conn),
	)

	developerService := developer_service.NewDeveloperServiceImpl(
		repos.NewDevelopersRepoImpl(conn),
	)

	inspectionService := inspect_service.NewInspectionService(
		inspect_service.InspectionServiceOptions{
			Manifests:     manifest_service.NewManifestServiceImpl(),
			Registry:      registry_service.NewRegistryServiceImpl(registry_service.RegistryServiceOpts{}),
			Github:        github_service.NewGithubServiceImpl(github_service.GithubServiceOpts{}),
			StackExchange: stack_exchange.NewClient(stack_exchange.Options{}),
			Developers:    developerService,
			Repo:          repos.NewInspectionsRepoImpl(conn),
		},
	)

	if config.Server.Enabled {
		err = databaseUpdater.Initialize(rootCtx)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize database updater: %s", err.Error()))
		}

		srv, err := cmd.CreateServer(cmd.ServerOptions{
			VulnerabilityService:         vulnerabilityService,
			VulnerabilityDatabaseService: databaseUpdater,
			InspectionService:            inspectionService,
			DeveloperService:             developerService,
			Config:                       &config.Server,
		})

		if err != nil {
			slog.ErrorContext(rootCtx, err.Error())
			os.Exit(1)
		}

		err = srv.Start(rootCtx)
		if err != nil {
			slog.ErrorContext(rootCtx, err.Error())
			os.Exit(1)
		}

		slog.InfoContext(rootCtx, fmt.Sprintf("server successfully started on http://%s:%d", config.Server.Host, config.Server.Port))

		select {
		case <-rootCtx.Done():
			err = srv.GracefulStop(rootCtx)
			if err != nil {
				slog.ErrorContext(rootCtx, "failed to gracefully stop grpc server: "+err.Error())
				os.Exit(1)
			}

			slog.WarnContext(rootCtx, "gracefully stopped grpc server, see you next time!")
			os.Exit(0)
		}
	}

	if !config.HasScan() {
		var err error
		config.Scan, err = cfg.NewScanFromArgs(rootCtx)

		if err != nil {
			slog.ErrorContext(rootCtx, fmt.Sprintf("failed to determine scan target(s): %s", err.Error()))
			os.Exit(1)
		}
	}

	err = cmd.RunCLI(rootCtx)
	if err != nil {
		return
	}
}
