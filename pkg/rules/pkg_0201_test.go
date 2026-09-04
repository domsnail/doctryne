package rules

import (
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

//go:embed pkg_0201_test.yaml
var testDataYaml_pkg_0201 []byte

//go:embed pkg_0201_test.json
var testDataJson_pkg_0201 []byte

func TestRule_PKG_0201(t *testing.T) {
	var rule Rule_PKG_0201

	t.Run("test_json_rule_parsing", func(t *testing.T) {
		var jsonRule Rule_PKG_0201
		var yamlRule Rule_PKG_0201

		err := json.Unmarshal(testDataJson_pkg_0201, &jsonRule)
		require.NoError(t, err)

		err = yaml.Unmarshal(testDataYaml_pkg_0201, &yamlRule)
		require.NoError(t, err)

		require.Equal(t, float64(72), jsonRule.Period.Hours())
		require.Equal(t, float64(72), yamlRule.Period.Hours())
		require.EqualValues(t, jsonRule, yamlRule)
		rule = yamlRule
	})

	t.Run("test_rule_preparation", func(t *testing.T) {
		err := rule.Prepare()
		require.NoError(t, err)
		require.NotNil(t, rule.Exec)
	})

	t.Run("test_rule_execution", func(t *testing.T) {
		ok, violation := rule.Exec(time.Now().Add(-4 * time.Hour))
		require.False(t, ok)
		require.Equal(t, "required amount of time has not been passed: time left 68h0m0s", violation)

		ok, violation = rule.Exec(time.Now().Add(4 * time.Hour))
		require.False(t, ok)
		require.Equal(t, "required amount of time has not been passed: time left 76h0m0s", violation)

		ok, violation = rule.Exec(time.Now().Add(-96 * time.Hour))
		require.True(t, ok)
		require.Empty(t, violation)
	})
}
