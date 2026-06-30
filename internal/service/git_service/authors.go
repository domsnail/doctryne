package git_service

import "github.com/domsnail/doctryne/internal/entity"

type authorsStore struct {
	authors []*entity.Developer

	authorsByEmail map[string]*entity.Developer
	authorsByName  map[string]*entity.Developer
}

func newAuthorsStore() *authorsStore {
	return &authorsStore{
		authors:        []*entity.Developer{},
		authorsByEmail: make(map[string]*entity.Developer),
		authorsByName:  make(map[string]*entity.Developer),
	}
}

func (store *authorsStore) get(email, name string) *entity.Developer {
	v, ok := store.authorsByEmail[email]
	if ok {
		return v
	}

	v, ok = store.authorsByName[name]
	if ok {
		return v
	}

	return nil
}

func (store *authorsStore) set(email, name string) *entity.Developer {
	var dev = entity.Developer{
		Name: name,
	}

	if email != "" {
		dev.Emails = []string{email}
		store.authorsByEmail[email] = &dev
	}

	store.authorsByName[name] = &dev
	store.authors = append(store.authors, &dev)

	return &dev
}

func (store *authorsStore) Update(email, name string) *entity.Developer {
	dev := store.get(email, name)
	if dev == nil {
		return store.set(email, name)
	}

	if dev.Name != name {
		store.authorsByName[name] = dev
	}

	if email != "" {
		dev.Emails = []string{email}
		store.authorsByEmail[email] = dev
	}

	return dev
}
