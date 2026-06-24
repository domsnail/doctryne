package inspect_service

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
)

type PackageInspectionPool struct {
	registry service.IRegistryService

	wg  *sync.WaitGroup
	ctx context.Context

	capacity int32
	active   atomic.Int32

	c *sync.Cond
}

func NewPackageInspectionPool(ctx context.Context, registry service.IRegistryService) *PackageInspectionPool {
	var mu sync.Mutex

	wg := &sync.WaitGroup{}

	return &PackageInspectionPool{
		registry: registry,
		capacity: cfg.GlobalConfig.Concurrency,
		wg:       wg,
		ctx:      ctx,
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

	pool.wg.Go(func() {
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
			slog.WarnContext(pool.ctx, "failed to inspect package",
				slog.String("package_name", pkg.Name),
				slog.String("package_version", pkg.Version),
				slog.String("error", err.Error()),
			)
		}

		// do not query GitHub or other vcs because some packages can be from the same git repository
	})
}

func (pool *PackageInspectionPool) Wait() error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		pool.wg.Wait()
	}()

	select {
	case <-pool.ctx.Done():
		return pool.ctx.Err()
	case <-done:
		return nil
	}
}
