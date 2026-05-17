package npm

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"time"
)

// Package is what is resolved from registry package page,
// ref: https://github.com/npm/registry/blob/main/docs/REGISTRY-API.md#package,
// for example: https://registry.npmjs.org/detect-libc
type Package struct {
	Id          string   `json:"_id"`
	Rev         string   `json:"_rev"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Homepage    string   `json:"homepage"`
	Keywords    []string `json:"keywords"`

	Versions map[string]Version `json:"versions"`

	DistTags struct {
		Next   string `json:"next"`
		Latest string `json:"latest"`
	} `json:"dist-tags"`

	Bugs struct {
		Url string `json:"url"`
	} `json:"bugs"`

	Author Person `json:"author"`

	Repository Repository `json:"repository"`

	Contributors []Person `json:"contributors"`
	Maintainers  []Person `json:"maintainers"`

	Readme         string `json:"readme"`
	ReadmeFilename string `json:"readmeFilename"`

	License string `json:"license"`

	Time map[string]time.Time `json:"time"`
}

// GetUniqueURLs extracts all unique url from all sources including:
// bugs, homepage, repository url
// Also, recursively calls the same method on all versions
func (pkg Package) GetUniqueURLs() map[string]URLAffiliation {
	var urls = make(map[string]URLAffiliation)

	// iterate over each available version
	for _, v := range pkg.Versions {
		maps.Copy(urls, v.GetUniqueURLs())
	}

	// top values should override version values
	if pkg.Bugs.Url != "" {
		urls[pkg.Bugs.Url] = URLAffiliation_Bugs
	}

	if pkg.Homepage != "" {
		urls[pkg.Homepage] = URLAffiliation_Homepage
	}

	if pkg.Repository.Url != "" {
		urls[pkg.Repository.Url] = URLAffiliation_Repository
	}

	slog.Debug(fmt.Sprintf("found %d unique urls", len(urls)),
		slog.String("package_name", pkg.Name),
	)

	return urls
}

// GetUniqueEmails extracts all unique emails from all actors including:
// contributors, authors, maintainers with their affiliation.
// Also, recursively calls the same method on all versions
func (pkg Package) GetUniqueEmails() map[string]PersonAffiliation {
	var emails = make(map[string]PersonAffiliation)

	// iterate over each available version
	for _, v := range pkg.Versions {
		maps.Copy(emails, v.GetUniqueEmails())
	}

	if pkg.Author.Email != "" {
		emails[pkg.Author.Email] = PersonAffiliation_Author
	}

	for _, c := range pkg.Contributors {
		emails[c.Email] = PersonAffiliation_Contributor
	}

	for _, m := range pkg.Maintainers {
		emails[m.Email] = PersonAffiliation_Maintainer
	}

	slog.Debug(fmt.Sprintf("found %d unique emails", len(emails)),
		slog.String("package_name", pkg.Name),
	)

	return emails
}

func (pkg Package) GetCreatedAt() (time.Time, error) {
	at, ok := pkg.Time["created"]
	if !ok {
		return time.Time{}, errors.New("no creation date found")
	}

	return at, nil
}

func (pkg Package) GetModifiedAt() (time.Time, error) {
	at, ok := pkg.Time["modified"]
	if !ok {
		return time.Time{}, errors.New("no modification date found")
	}

	return at, nil
}
