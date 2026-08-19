package inspect_service

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/domsnail/doctryne/cfg"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/service"
	"github.com/domsnail/doctryne/pkg/stack_exchange"
)

type DeveloperInspectionPool struct {
	github        service.IGithubService
	stackExchange *stack_exchange.Client

	wg  *sync.WaitGroup
	ctx context.Context

	capacity int32
	active   atomic.Int32

	totalInspections      atomic.Int32
	successfulInspections atomic.Int32

	c *sync.Cond
}

func NewDeveloperInspectionPool(ctx context.Context, github service.IGithubService, stackExchange *stack_exchange.Client) *DeveloperInspectionPool {
	var mu sync.Mutex

	wg := &sync.WaitGroup{}

	return &DeveloperInspectionPool{
		github:        github,
		stackExchange: stackExchange,
		wg:            wg,
		ctx:           ctx,
		capacity:      cfg.GlobalConfig.Concurrency,
		c:             sync.NewCond(&mu),
	}
}

type InspectionSource int

const (
	InspectionSource_GitHub InspectionSource = iota
	InspectionSource_StackExchange

	InspectionSource_DevTo
	InspectionSource_HeadHunter
	InspectionSource_Telegram
)

func (pool *DeveloperInspectionPool) Inspect(developer *entity.Developer, source InspectionSource) error {
	if developer == nil {
		return errors.New("developer is nil")
	} else if developer.Username == "" {
		slog.WarnContext(pool.ctx, "skipping developer inspection: no username provided",
			slog.String("name", developer.Name),
			slog.String("info_source", source.String()),
		)

		return errors.New("no developer username provided")
	}

	pool.c.L.Lock()
	for pool.active.Load() >= pool.capacity {
		pool.c.Wait()
	}

	pool.active.Add(1)
	pool.c.L.Unlock()

	pool.wg.Go(func() {
		defer func() {
			developer.LastLookupAt = new(time.Now())

			pool.c.L.Lock()
			pool.active.Add(-1)
			pool.c.L.Unlock()

			pool.c.Signal()
			pool.totalInspections.Add(1)
		}()

		slog.DebugContext(pool.ctx, "inspecting developer...",
			slog.String("username", developer.Username),
			slog.String("info_source", source.String()),
		)

		var err error

		switch source {
		case InspectionSource_GitHub:
			err = pool.inspectGitHub(pool.ctx, developer)
		case InspectionSource_StackExchange:
			err = pool.inspectStackExchange(pool.ctx, developer)
		default:
			slog.ErrorContext(pool.ctx, "developer info source is not supported",
				slog.String("info_source", source.String()),
			)
		}

		if err != nil {
			slog.WarnContext(pool.ctx, "failed to inspect developer profile",
				slog.String("info_source", source.String()),
				slog.String("error", err.Error()),
			)
		} else {
			pool.successfulInspections.Add(1)
		}
	})

	return nil
}

func (pool *DeveloperInspectionPool) inspectGitHub(ctx context.Context, developer *entity.Developer) (err error) {
	var profile *entity.GithubDeveloperProfile

	if developer.GithubID != nil {
		profile, err = pool.github.GetProfileByID(ctx, *developer.GithubID)
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
	developer.GithubProfile = profile

	slog.DebugContext(pool.ctx, "found developer github profile",
		slog.String("username", developer.Username),
		slog.Int64("github_profile", *developer.GithubID),
	)

	return nil
}

func (pool *DeveloperInspectionPool) inspectStackExchange(ctx context.Context, developer *entity.Developer) (err error) {
	profiles, _, err := pool.stackExchange.GetUsersByUsername(ctx, developer.Username)
	if err != nil {
		return err
	} else if profiles == nil || len(profiles) == 0 {
		return errors.New("developer stack exchange profile not found")
	}

	var profile *stack_exchange.User
	if len(profiles) > 1 {
		slog.DebugContext(pool.ctx, "multiple developer stack exchange profiles found, trying to find exact match",
			slog.String("username", developer.Username),
			slog.Int("profiles_found", len(profiles)),
		)

		i := slices.IndexFunc(profiles, func(p *stack_exchange.User) bool {
			return strings.EqualFold(p.DisplayName, developer.Username) || strings.EqualFold(p.DisplayName, developer.Name)
		})

		if i == -1 {
			return errors.New("multiple developer stack exchange profiles found, count not find exact match")
		}

		profile = profiles[i]
	} else {
		profile = profiles[0]
	}

	developer.StackExchangeAccountID = profile.AccountID
	developer.StackExchangeProfile = &entity.StackExchangeDeveloperProfile{
		UserID:       profile.UserID,
		AccountID:    profile.AccountID,
		DisplayName:  profile.DisplayName,
		WebsiteUrl:   profile.WebsiteUrl,
		AboutMe:      profile.AboutMe,
		Location:     profile.Location,
		IsEmployee:   profile.IsEmployee,
		IsRegistered: profile.IsRegistered(),
		Reputation:   profile.Reputation,
		Badges: entity.StackExchangeBadges{
			Bronze: profile.BadgeCounts.Bronze,
			Silver: profile.BadgeCounts.Silver,
			Gold:   profile.BadgeCounts.Gold,
		},
		CreatedAt:    profile.CreationDate.Time,
		LastAccessAt: profile.LastAccessDate.Time,
	}

	if !profile.LastModifiedDate.Time.IsZero() {
		developer.StackExchangeProfile.UpdatedAt = &profile.LastModifiedDate.Time
	}

	if !profile.TimedPenaltyDate.Time.IsZero() {
		developer.StackExchangeProfile.PenaltyTill = &profile.TimedPenaltyDate.Time
	}

	slog.DebugContext(pool.ctx, "found developer stack exchange profile",
		slog.String("username", developer.Username),
		slog.Uint64("account_id", *developer.StackExchangeAccountID),
	)

	return nil
}

func (pool *DeveloperInspectionPool) stats() (total, successes int32) {
	return pool.totalInspections.Load(), pool.successfulInspections.Load()
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
