package types

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// Semver implements https://semver.org/
type Semver struct {
	Major, Minor, Patch, Revision semverPart

	PrereleaseIdentifiers []string
	BuildMetadata         []string

	raw string
}

func (semver *Semver) Parse(s string) error {
	if s == "" || s == "n/a" {
		semver.Major, semver.Minor, semver.Patch = semverPart{value: 0}, semverPart{value: 0}, semverPart{value: 0}
		semver.raw = "0"

		return nil
	}

	semver.raw = s
	s = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(s), "v"))
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i]
	}

	var pre string
	if i := strings.IndexByte(s, '-'); i >= 0 {
		s, pre = s[:i], s[i+1:]
		if pre == "" {
			return fmt.Errorf("empty prerelease identifier provided for version: '%s'", s)
		}
	}

	var build string
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s, build = s[:i], s[i+1:]
		if pre == "" {
			return fmt.Errorf("empty build metadata provided for version: '%s'", s)
		}
	}

	parts := strings.SplitN(s, ".", 4)
	if len(parts) == 0 {
		return fmt.Errorf("expected 4 parts (major.minor.patch.revision) for version: '%s'", s)
	}

	var nums [4]semverPart
	for i, p := range parts {
		if p == "*" || strings.EqualFold(p, "x") {
			// missing parts after a wildcard are wildcards too: 8.* == 8.*.*
			for j := i; j < len(nums); j++ {
				nums[j] = semverPart{wildcard: true}
			}

			break
		}

		n, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			if i == 3 {
				// cases: 3.0.0.beta (GHSA-92w9-2pqw-rhjj, ruby on rails)
				pre += p
				continue
			}

			return fmt.Errorf("semver part '%q' is not a number, version %s: %w", p, s, err)
		}

		nums[i] = semverPart{value: n}
	}

	semver.Major, semver.Minor, semver.Patch, semver.Revision = nums[0], nums[1], nums[2], nums[3]

	if pre != "" {
		semver.PrereleaseIdentifiers = strings.Split(pre, ".")
		if slices.Contains(semver.PrereleaseIdentifiers, "") {
			return fmt.Errorf("empty prerelease identifier in version: %s", s)
		}
	}

	if build != "" {
		semver.BuildMetadata = strings.Split(build, ".")
		if slices.Contains(semver.BuildMetadata, "") {
			return fmt.Errorf("empty build metadata in version: %s", s)
		}
	}

	return nil
}

func (semver *Semver) Compare(to Semver) int {
	if r := semver.Major.compare(to.Major); r != 0 {
		return r
	}

	if r := semver.Minor.compare(to.Minor); r != 0 {
		return r
	}

	if r := semver.Patch.compare(to.Patch); r != 0 {
		return r
	}

	if r := semver.Revision.compare(to.Revision); r != 0 {
		return r
	}

	// 3.0.0-rc1 < 3.0.0.
	switch {
	case len(semver.PrereleaseIdentifiers) == 0 && len(to.PrereleaseIdentifiers) == 0:
		return 0
	case len(semver.PrereleaseIdentifiers) == 0:
		return 1
	case len(to.PrereleaseIdentifiers) == 0:
		return -1
	}

	for i := 0; i < len(semver.PrereleaseIdentifiers) && i < len(to.PrereleaseIdentifiers); i++ {
		a, b := semver.PrereleaseIdentifiers[i], to.PrereleaseIdentifiers[i]
		an, aErr := strconv.ParseUint(a, 10, 64)
		bn, bErr := strconv.ParseUint(b, 10, 64)

		switch {
		case aErr == nil && bErr == nil:
			if r := cmp.Compare(an, bn); r != 0 {
				return r
			}
		case aErr == nil: // numeric identifiers have lower precedence: https://semver.org/#spec-item-11
			return -1
		case bErr == nil:
			return 1
		default:
			if r := cmp.Compare(a, b); r != 0 {
				return r
			}
		}
	}

	// do not compare by build metadata, see: https://semver.org/#spec-item-10

	// 3.0.0-rc < 3.0.0-rc.1
	return cmp.Compare(len(semver.PrereleaseIdentifiers), len(to.PrereleaseIdentifiers))
}

// HasWildcard reports whether any of major, minor or patch is '*'.
func (semver *Semver) HasWildcard() bool {
	return semver.Major.wildcard || semver.Minor.wildcard || semver.Patch.wildcard
}

func (semver *Semver) String() string {
	return semver.raw
}

type semverPart struct {
	value    uint64
	wildcard bool
}

func (p semverPart) compare(to semverPart) int {
	if p.wildcard || to.wildcard {
		return 0
	}

	return cmp.Compare(p.value, to.value)
}
