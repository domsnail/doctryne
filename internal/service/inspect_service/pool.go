package inspect_service

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"golang.org/x/sync/errgroup"
)

type PackageInspectionPool struct {
	errGroup *errgroup.Group
	ctx      context.Context

	capacity int32
	active   atomic.Int32

	c *sync.Cond
}

func NewPackageInspectionPool(ctx context.Context) *PackageInspectionPool {
	var mu sync.Mutex

	group, groupCtx := errgroup.WithContext(ctx)

	return &PackageInspectionPool{
		capacity: cfg.GlobalConfig.Concurrency,
		errGroup: group,
		ctx:      groupCtx,
		c:        sync.NewCond(&mu),
	}
}

func (pool *PackageInspectionPool) Inspect(pkg *entity.Package) {
	defer pool.c.L.Unlock()
	pool.c.L.Lock()

	for pool.active.Load() >= pool.capacity {
		pool.c.Wait()
	}

	pool.active.Add(1)

	pool.errGroup.Go(func() error {
		defer func() {
			pool.active.Add(-1)
			pool.c.Signal()
		}()

		slog.DebugContext(pool.ctx, "inspecting package...",
			slog.String("package_name", pkg.Name),
			slog.String("package_version", pkg.Version),
		)

		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)

		return nil
	})
}

func (pool *PackageInspectionPool) Wait() error {
	return pool.errGroup.Wait()
}
