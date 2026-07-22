package inspect_service

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
)

type DeveloperInspectionPool struct {
	github service.IGithubService

	wg  *sync.WaitGroup
	ctx context.Context

	capacity int32
	active   atomic.Int32

	c *sync.Cond
}

type InspectionSource int

const (
	InspectionSource_GitHub InspectionSource = iota
	InspectionSource_StackExchange

	InspectionSource_DevTo
	InspectionSource_HeadHunter
	InspectionSource_Telegram
)

func (pool *DeveloperInspectionPool) Inspect(ctx context.Context, developer *entity.Developer, source InspectionSource) {
	if developer == nil {
		return
	} else if developer.Username == "" {
		slog.WarnContext(ctx, "skipping developer inspection: no username provided")
		return
	}

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

		slog.DebugContext(pool.ctx, "inspecting developer...",
			slog.String("username", developer.Username),
			slog.String("info_source", source.String()),
		)

		var err error

		switch source {
		case InspectionSource_GitHub:
			err = pool.inspectGitHub(ctx, developer)
		default:
			slog.ErrorContext(pool.ctx, "developer info source is not supported",
				slog.String("info_source", source.String()),
			)
		}

		if err != nil {
			slog.WarnContext(pool.ctx, "error inspecting developer profile",
				slog.String("info_source", source.String()),
				slog.String("error", err.Error()),
			)
		}
	})

	return
}

func (pool *DeveloperInspectionPool) inspectGitHub(ctx context.Context, developer *entity.Developer) (err error) {
	if developer.GithubMetadata != nil {
		slog.WarnContext(ctx, "skipping developer github profile inspection: metadata already exists")
		return nil
	}

	var profile *entity.GithubDeveloperProfile

	if developer.GithubID != 0 {
		profile, err = pool.github.GetProfileByID(ctx, developer.GithubID)
		if err != nil {
			return err
		}
	} else {
		profile, err = pool.github.GetProfileByUsername(ctx, developer.Username)
		if err != nil {
			return err
		}
	}

	if profile == nil {
		return errors.New("developer github profile not found")
	}

	developer.GithubID = profile.ID
	developer.GithubMetadata = profile
	return nil
}

func (source InspectionSource) String() string {
	switch source {
	case InspectionSource_GitHub:
		return "github"
	case InspectionSource_StackExchange:
		return "stack_exchange"
	case InspectionSource_DevTo:
		return "dev_to"
	case InspectionSource_HeadHunter:
		return "head_hunter"
	case InspectionSource_Telegram:
		return "telegram"
	default:
		return "unspecified"
	}
}
