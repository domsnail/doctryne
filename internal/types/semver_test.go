package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Semver(t *testing.T) {
	t.Run("test semver parsing", func(t *testing.T) {
		var cases = []struct {
			version string
			ok      bool
		}{
			{version: "0.0.0", ok: true},
			{version: "0.1.0", ok: true},
			{version: "0.2.0", ok: true},
			{version: "0.2.0-beta", ok: true},
			{version: "0.2.0-beta2", ok: true},
			{version: "3.0.0-beta4", ok: true},
			{version: "3.0.0-beta3+build.7", ok: true},
			{version: "3.0.0.1-beta3+build.7", ok: false},
			{version: "pre3.0.0.1-beta3+build.7", ok: false},
			{version: "pre3.0.0.1beta3+build.7", ok: false},
		}

		for _, c := range cases {
			v := &Semver{}
			err := v.Parse(c.version)

			t.Logf("semver: %d.%d.%d (%v), err: %v", v.Major, v.Minor, v.Patch, v.PrereleaseIdentifiers, err)
			require.Equal(t, c.ok, err == nil)
		}
	})

	t.Run("test semver ranges", func(t *testing.T) {
		var target = "2.1.0-beta1+build.7"
		cv := &Semver{}
		err := cv.Parse(target)
		require.NoError(t, err)

		var cases = []struct {
			version string
			must    int
		}{
			{version: "0.0.0", must: 1},
			{version: "0.1.0", must: 1},
			{version: "v0.1.12", must: 1},
			{version: "0.2.0", must: 1},
			{version: "0.2.0-beta", must: 1},
			{version: "v0.2.0-beta", must: 1},
			{version: "1.2.0-beta2", must: 1},
			{version: "2.0.0-beta4", must: 1},
			{version: "2.1.0-beta1+build.7", must: 0},
			{version: "2.1.0-beta1+build.12", must: 0},
			{version: "v2.1.0-beta1+build.12", must: 0},
			{version: "2.1.0-beta2", must: -1},
			{version: "2.1.0", must: -1},
			{version: "v2.1.0", must: -1},
			{version: "3.0.0-beta3", must: -1},
			{version: "4.0.1-beta3+build.7", must: -1},
			{version: "v5.0.1-beta3+build.7", must: -1},
		}

		for _, c := range cases {
			v := Semver{}
			err = v.Parse(c.version)
			require.NoError(t, err)

			res := cv.Compare(v)
			t.Logf("compare %s to %s, result: %d", cv.raw, v.raw, res)
			require.EqualValues(t, c.must, res)
		}
	})

	t.Run("test semver ranges as numeric", func(t *testing.T) {
		var target = "6.2.0-beta.12"
		cv := &Semver{}
		err := cv.Parse(target)
		require.NoError(t, err)

		var cases = []struct {
			version string
			must    int
		}{
			{version: "0.0.0", must: 1},
			{version: "0.1.0", must: 1},
			{version: "v0.1.12", must: 1},
			{version: "0.2.0", must: 1},
			{version: "0.2.0-beta", must: 1},
			{version: "v0.2.0-beta", must: 1},
			{version: "1.2.0-beta2", must: 1},
			{version: "2.0.0-beta4", must: 1},
			{version: "2.1.0-beta1+build.7", must: 1},
			{version: "2.1.0-beta1+build.12", must: 1},
			{version: "v2.1.0-beta1+build.12", must: 1},
			{version: "6.2.0-alpha.1", must: 1},
			{version: "6.2.0-alpha.2", must: 1},
			{version: "6.2.0-alpha.8.beta", must: 1},
			{version: "6.2.0-alpha.8.beta.2", must: 1},
			{version: "6.2.0-beta.1", must: 1},
			{version: "6.2.0-beta.7", must: 1},
			{version: "6.2.0-beta.10", must: 1},
			{version: "6.2.0-beta.12.alpha", must: 1},
			{version: "6.2.0-beta.12.alpha.2", must: 1},
			{version: "6.2.0-beta.15", must: -1},
			{version: "6.2.0-beta.21", must: -1},
			{version: "9.1.0", must: -1},
		}

		for _, c := range cases {
			v := Semver{}
			err = v.Parse(c.version)
			require.NoError(t, err)

			res := cv.Compare(v)
			t.Logf("compare %s to %s, result: %d (must be: %d)", cv.raw, v.raw, res, c.must)
			require.EqualValues(t, c.must, res)
		}
	})
}
