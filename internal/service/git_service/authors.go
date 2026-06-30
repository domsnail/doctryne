package git_service

import (
	"slices"

	"github.com/domsnail/doctryne/internal/entity"
)

type authorsStore struct {
	authors []*entity.Developer

	authorsByEmail map[string]*entity.Developer
	authorsByName  map[string]*entity.Developer

	authorStats map[*entity.Developer]*entity.CommitStats
}

func newAuthorsStore() *authorsStore {
	return &authorsStore{
		authors:        []*entity.Developer{},
		authorsByEmail: make(map[string]*entity.Developer),
		authorsByName:  make(map[string]*entity.Developer),
		authorStats:    make(map[*entity.Developer]*entity.CommitStats),
	}
}

func (store *authorsStore) get(name string) *entity.Developer {
	v, ok := store.authorsByName[name]
	if ok {
		return v
	}

	return nil
}

func (store *authorsStore) set(name, email string) *entity.Developer {
	var dev = entity.Developer{
		Name: name,
	}

	if email != "" {
		dev.Emails = []string{email}
		store.authorsByEmail[email] = &dev
	}

	store.authorsByName[name] = &dev
	store.authors = append(store.authors, &dev)
	store.authorStats[&dev] = new(entity.CommitStats)

	return &dev
}

func (store *authorsStore) Update(email, name string) *entity.Developer {
	dev := store.get(name)
	if dev == nil {
		return store.set(name, email)
	}

	if email != "" && !slices.Contains(dev.Emails, email) {
		dev.Emails = append(dev.Emails, email)
		store.authorsByEmail[email] = dev
	}

	return dev
}

func (store *authorsStore) AddStats(developer *entity.Developer, stats *entity.CommitStats) {
	store.authorStats[developer].Add(stats)
	return
}

func (store *authorsStore) Stats() entity.DeveloperCommitStats {
	return store.authorStats
}
