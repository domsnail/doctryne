package rules

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

//go:embed cve_0101_test.yaml
var testDataYaml_cve_0101 []byte

//go:embed cve_0101_test.json
var testDataJson_cve_0101 []byte

func TestRule_CVE_0101(t *testing.T) {
	var maxRule Rule_CVE_0101
	var minMaxRule Rule_CVE_0101

	t.Run("test_json_rule_parsing", func(t *testing.T) {
		var jsonRule Rule_CVE_0101
		var yamlRule Rule_CVE_0101

		err := json.Unmarshal(testDataJson_cve_0101, &jsonRule)
		require.NoError(t, err)

		err = yaml.Unmarshal(testDataYaml_cve_0101, &yamlRule)
		require.NoError(t, err)

		require.Equal(t, float32(9.0), *yamlRule.Max)
		maxRule = yamlRule

		require.Equal(t, float32(9.0), *jsonRule.Max)
		require.Equal(t, float32(2.0), *jsonRule.Min)
		minMaxRule = jsonRule
	})

	t.Run("test_rules_preparation", func(t *testing.T) {
		err := maxRule.Prepare()
		require.NoError(t, err)
		require.NotNil(t, maxRule.Exec)

		err = minMaxRule.Prepare()
		require.NoError(t, err)
		require.NotNil(t, minMaxRule.Exec)
	})

	t.Run("test_max_rule_execution", func(t *testing.T) {
		ok, violation := maxRule.Exec(9.3)
		require.False(t, ok)
		require.Equal(t, "value out of range", violation)

		ok, violation = maxRule.Exec(6)
		require.True(t, ok)
		require.Empty(t, violation)

		ok, violation = maxRule.Exec(1)
		require.True(t, ok)
		require.Empty(t, violation)
	})

	t.Run("test_min_max_rule_execution", func(t *testing.T) {
		ok, violation := minMaxRule.Exec(9.3)
		require.False(t, ok)
		require.Equal(t, "value out of range", violation)

		ok, violation = minMaxRule.Exec(6)
		require.True(t, ok)
		require.Empty(t, violation)

		ok, violation = minMaxRule.Exec(1)
		require.False(t, ok)
		require.Equal(t, "value out of range", violation)
	})
}
