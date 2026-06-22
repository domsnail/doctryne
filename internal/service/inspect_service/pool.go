package inspect_service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
	"golang.org/x/sync/errgroup"
)

type PackageInspectionPool struct {
	registry service.IRegistryService

	errGroup *errgroup.Group
	ctx      context.Context

	capacity int32
	active   atomic.Int32

	c *sync.Cond
}

func NewPackageInspectionPool(ctx context.Context, registry service.IRegistryService) *PackageInspectionPool {
	var mu sync.Mutex

	group, groupCtx := errgroup.WithContext(ctx)

	return &PackageInspectionPool{
		registry: registry,
		capacity: cfg.GlobalConfig.Concurrency,
		errGroup: group,
		ctx:      groupCtx,
		c:        sync.NewCond(&mu),
	}
}

func (pool *PackageInspectionPool) Inspect(pkg *entity.Package) {
	pool.c.L.Lock()
	for pool.active.Load() >= pool.capacity {
		pool.c.Wait()
	}

	pool.active.Add(1)
	pool.c.L.Unlock()

	pool.errGroup.Go(func() error {
		defer func() {
			pool.c.L.Lock()
			pool.active.Add(-1)
			pool.c.L.Unlock()

			pool.c.Signal()
		}()

		slog.DebugContext(pool.ctx, "inspecting package...",
			slog.String("package_name", pkg.Name),
			slog.String("package_version", pkg.Version),
		)

		err := pool.registry.GetPackageInfo(pool.ctx, pkg)
		if err != nil {
			return fmt.Errorf("failed to get package info: %w", err)
		}

		return nil
	})
}

func (pool *PackageInspectionPool) Wait() error {
	return pool.errGroup.Wait()
}
