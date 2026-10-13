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
	Major, Minor, Patch   uint64
	PrereleaseIdentifiers []string

	raw string
}

func (semver *Semver) Parse(s string) error {
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
		// try to compare as numbers: https://semver.org/
		r, err := compareAsNumbers(semver.PrereleaseIdentifiers[i], to.PrereleaseIdentifiers[i])
		if err == nil && r != 0 {
			return r
		}

		// cmp compares as strings
		if r = cmp.Compare(semver.PrereleaseIdentifiers[i], to.PrereleaseIdentifiers[i]); r != 0 {
			return r
		}
	}

	return cmp.Compare(len(to.PrereleaseIdentifiers), len(semver.PrereleaseIdentifiers))
}

func compareAsNumbers(a, b string) (int, error) {
	ident1, err := strconv.ParseUint(a, 10, 64)
	if err != nil {
		return 0, err
	}

	ident2, err := strconv.ParseUint(b, 10, 64)
	if err != nil {
		return 0, err
	}

	return cmp.Compare(ident1, ident2), nil
}

func (semver *Semver) String() string {
	return semver.raw
}
