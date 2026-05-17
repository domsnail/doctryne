package npm

// ref: https://github.com/npm/registry/blob/main/docs/REGISTRY-API.md#version
type Version struct {
	Id         string `json:"_id"`
	From       string `json:"_from"`
	SHA        string `json:"_shasum"`
	NpmVersion string `json:"_npmVersion"`

	Homepage    string   `json:"homepage"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Main        string   `json:"main"`
	Keywords    []string `json:"keywords"`

	Author Person `json:"author"`

	License string `json:"license"`

	Maintainers []Person   `json:"maintainers"`
	NpmUser     Person     `json:"_npmUser"`
	Repository  Repository `json:"repository"`

	Bugs struct {
		Url string `json:"url"`
	} `json:"bugs"`

	GitHead     string `json:"gitHead"`
	NodeVersion string `json:"_nodeVersion"`

	Bin map[string]string `json:"bin"`

	Dist struct {
		SHA       string `json:"shasum"`
		Tarball   string `json:"tarball"`
		Integrity string `json:"integrity"`

		Signatures []Signature `json:"signatures"`
	} `json:"dist"`

	Engines struct {
		Node string `json:"node"`
	} `json:"engines"`

	Scripts map[string]string `json:"scripts"`

	Directories struct {
	} `json:"directories"`

	DevDependencies map[string]string `json:"devDependencies"`

	NpmOperationalInternal struct {
		Tmp  string `json:"tmp"`
		Host string `json:"host"`
	} `json:"_npmOperationalInternal"`
}

// GetUniqueEmails extracts all unique emails from all actors including:
// author, maintainers, npm user with their affiliation
func (ver Version) GetUniqueEmails() map[string]PersonAffiliation {
	var emails = make(map[string]PersonAffiliation)

	if ver.Author.Email != "" {
		emails[ver.Author.Email] = PersonAffiliation_Author
	}

	if ver.NpmUser.Email != "" {
		emails[ver.NpmUser.Email] = PersonAffiliation_Author
	}

	for _, m := range ver.Maintainers {
		emails[m.Email] = PersonAffiliation_Maintainer
	}

	return emails
}

// GetUniqueURLs extracts all unique url from all sources including:
// bugs, homepage, repository url, registry (tarball)
// Also, recursively calls the same method on all versions
func (ver Version) GetUniqueURLs() map[string]URLAffiliation {
	var urls = make(map[string]URLAffiliation)

	if ver.Bugs.Url != "" {
		urls[ver.Bugs.Url] = URLAffiliation_Bugs
	}

	if ver.Homepage != "" {
		urls[ver.Homepage] = URLAffiliation_Homepage
	}

	if ver.Repository.Url != "" {
		urls[ver.Repository.Url] = URLAffiliation_Repository
	}

	if ver.Dist.Tarball != "" {
		urls[ver.Dist.Tarball] = URLAffiliation_Registry
	}

	return urls
}
