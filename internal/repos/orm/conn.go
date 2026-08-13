package orm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/domsnail/doctryne/cfg"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDatabaseConn(ctx context.Context, config cfg.DatabaseConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch config.Driver {
	case "postgres":
		dialector = postgres.Open(newPostgresConnectionString(config))
	case "sqlite3", "sqlite":
		dialector = sqlite.Open(newSqliteConnectionString(config))
	default:
		return nil, errors.New(fmt.Sprintf("unsupported database driver: %s", config.Driver))
	}

	slog.DebugContext(ctx, "connecting to database...",
		slog.String("driver", config.Driver),
		slog.String("host", net.JoinHostPort(config.Host, strconv.Itoa(int(config.Port)))),
		slog.String("database", config.Name),
		slog.Bool("ssl_mode", config.Ssl),
	)

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                    newGormSlogger(config),
		DefaultTransactionTimeout: time.Second * 120,
		DefaultContextTimeout:     time.Second * 30,
		FullSaveAssociations:      true,
	})

	if err != nil {
		slog.ErrorContext(ctx, "failed to connect to database",
			slog.String("driver", config.Driver),
			slog.String("host", net.JoinHostPort(config.Host, strconv.Itoa(int(config.Port)))),
			slog.String("file", config.File),
			slog.String("database", config.Name),
			slog.Bool("ssl_mode", config.Ssl),
			slog.String("error", err.Error()),
		)

		return nil, err
	}

	slog.InfoContext(ctx, "successfully connected to database",
		slog.String("driver", config.Driver),
		slog.String("host", net.JoinHostPort(config.Host, strconv.Itoa(int(config.Port)))),
		slog.String("file", config.File),
		slog.String("database", config.Name),
		slog.Bool("ssl_mode", config.Ssl),
	)

	return db, err
}

func newPostgresConnectionString(config cfg.DatabaseConfig) string {
	dsn := strings.Builder{}
	dsn.WriteString(fmt.Sprintf("host=%s ", config.Host))
	dsn.WriteString(fmt.Sprintf("port=%d ", config.Port))
	dsn.WriteString(fmt.Sprintf("user=%s ", config.User))

	if config.Pass != "" {
		dsn.WriteString(fmt.Sprintf("password=%s ", config.Pass))
	}

	if config.Name != "" {
		dsn.WriteString(fmt.Sprintf("dbname=%s ", config.Name))
	}

	if config.Ssl {
		dsn.WriteString("sslmode=enable ")
	} else {
		dsn.WriteString("sslmode=disable ")
	}

	if config.Timezone != "" {
		dsn.WriteString(fmt.Sprintf("TimeZone=%s ", config.Timezone))
	}

	return strings.TrimSpace(dsn.String())
}

func newSqliteConnectionString(config cfg.DatabaseConfig) string {
	return strings.TrimSpace(config.File)
}

func newGormSlogger(config cfg.DatabaseConfig) logger.Interface {
	var options = []slogGorm.Option{
		slogGorm.WithHandler(slog.Default().Handler()),
	}

	if config.TraceAll {
		options = append(options, slogGorm.WithTraceAll())
	} else {
		options = append(options, slogGorm.WithIgnoreTrace())
	}

	return slogGorm.New(options...)
}
