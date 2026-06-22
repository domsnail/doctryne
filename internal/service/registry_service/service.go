package registry_service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/pkg/npm"
	"github.com/domsnail/doctryne/pkg/types"
)

type RegistryServiceImpl struct {
	npm *npm.Client

	statsPeriod time.Duration
}

type RegistryServiceOpts struct {
	Timeout time.Duration

	BearerToken string
	ApiURL      string
	RegistryURL string

	LatestActivityPeriod time.Duration
}

const defaultLatestActivityPeriod = 30 * 24 * time.Hour

func NewRegistryServiceImpl(opts RegistryServiceOpts) *RegistryServiceImpl {
	npmClient, err := npm.NewClient(npm.Options{
		Timeout:     opts.Timeout,
		BearerToken: opts.BearerToken,
		ApiURL:      opts.ApiURL,
		RegistryURL: opts.RegistryURL,
	})

	if err != nil {
		slog.Error("failed to init npm client", slog.String("error", err.Error()))
		panic(err)
	}

	if opts.LatestActivityPeriod <= 0 {
		slog.Debug("latest activity is not set or invalid, setting default period...",
			slog.Duration("period", defaultLatestActivityPeriod),
		)

		opts.LatestActivityPeriod = defaultLatestActivityPeriod
	}

	impl := RegistryServiceImpl{
		npm:         npmClient,
		statsPeriod: opts.LatestActivityPeriod,
	}

	return &impl
}

func (service RegistryServiceImpl) GetPackageInfo(ctx context.Context, pkg *entity.Package) error {
	if pkg == nil {
		return errors.New("package is nil")
	}

	switch pkg.Ecosystem {
	case types.Ecosystem_NPM:
		err := service.queryNPM(ctx, pkg)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("package ecosystem '%s' not supported", pkg.Ecosystem)
	}

	return nil
}
