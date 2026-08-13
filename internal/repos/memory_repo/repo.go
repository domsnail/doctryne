package memory_repo

import (
	"context"
	"errors"
	"sync"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/pkg/types"
	"github.com/google/uuid"
)

type InMemoryRepoImpl struct {
	mu sync.RWMutex

	s map[string][]*entity.Inspection
}

func NewInMemoryRepo() *InMemoryRepoImpl {
	impl := InMemoryRepoImpl{
		mu: sync.RWMutex{},
		s:  make(map[string][]*entity.Inspection),
	}

	impl.s["af6f4cbd-df32-4ae8-ae28-8df1f41d5ffd"] = []*entity.Inspection{
		entity.NewInspection(&entity.InspectionOptions{
			ScanType:                    types.ScanType_Binary,
			ManifestType:                types.ManifestType_CycloneDX,
			Mode:                        types.InspectionMode_Direct,
			ExtractFullOrganizationInfo: false,
			ExtractFullContributorInfo:  false,
			DeepRepositoryInspection:    false,
			InspectIssues:               false,
		}),
	}

	return &impl
}

func (repo *InMemoryRepoImpl) CreateInspection(ctx context.Context, ins *entity.Inspection) error {
	if ins == nil {
		return errors.New("inspection is nil")
	}

	uuid_ := ins.UUID.String()
	if uuid_ == "" {
		return errors.New("inspection uuid is empty")
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	revisions, ok := repo.s[uuid_]
	if ok {
		ins.Revision = uint32(len(revisions)) + 1

		revisions = append(revisions, ins)
		repo.s[uuid_] = revisions

		return nil
	}

	ins.Revision = 1
	repo.s[uuid_] = []*entity.Inspection{ins}

	return nil
}

func (repo *InMemoryRepoImpl) SelectInspectionByUUID(ctx context.Context, uid uuid.UUID) (*entity.Inspection, error) {
	if uid.String() == "" {
		return nil, errors.New("uid is empty")
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	revisions, ok := repo.s[uid.String()]
	if !ok {
		return nil, errors.New("inspection not found")
	} else if len(revisions) == 0 {
		return nil, errors.New("no revisions found")
	}

	return revisions[len(revisions)-1], nil
}

func (repo *InMemoryRepoImpl) SelectInspectionRevisionByUUID(ctx context.Context, uid uuid.UUID, rev uint32) (*entity.Inspection, error) {
	if uid.String() == "" {
		return nil, errors.New("uid is empty")
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	revisions, ok := repo.s[uid.String()]
	if !ok {
		return nil, errors.New("inspection not found")
	}

	if rev > uint32(len(revisions)) {
		return nil, errors.New("revision out of range")
	}

	return revisions[rev-1], nil
}
