package types

import (
	"cmp"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
)

// Semver implements https://semver.org/
type Semver struct {
	Major, Minor, Patch   uint64
	PrereleaseIdentifiers []string

	raw string
}

func (semver *Semver) Parse(s string) error {
	if s == "" || s == "n/a" {
		slog.Debug("empty semver value provided", slog.String("value", s))
		semver.Major, semver.Minor, semver.Patch = 0, 0, 0
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
			return fmt.Errorf("empty prerelease identifier provided for version: %s", s)
		}
	}

	parts := strings.Split(s, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return fmt.Errorf("expected 3 parts (major.minor.patch) for version: %s", s)
	}

	var nums [3]uint64
	for i, p := range parts {
		n, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return fmt.Errorf("semver part '%q' is not a number, version %s: ", p, s)
		}

		nums[i] = n
	}

	semver.Major, semver.Minor, semver.Patch = nums[0], nums[1], nums[2]
	if pre != "" {
		semver.PrereleaseIdentifiers = strings.Split(pre, ".")
		if slices.Contains(semver.PrereleaseIdentifiers, "") {
			return fmt.Errorf("empty prerelease identifier in version: %s", s)
		}
	}

	return nil
}

func (semver *Semver) Compare(to Semver) int {
	if r := cmp.Compare(semver.Major, to.Major); r != 0 {
		return r
	}

	if r := cmp.Compare(semver.Minor, to.Minor); r != 0 {
		return r
	}

	if r := cmp.Compare(semver.Patch, to.Patch); r != 0 {
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

	// 3.0.0-rc < 3.0.0-rc.1
	return cmp.Compare(len(semver.PrereleaseIdentifiers), len(to.PrereleaseIdentifiers))
}

func (semver *Semver) String() string {
	return semver.raw
}
