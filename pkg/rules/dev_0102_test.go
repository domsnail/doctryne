package rules

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

//go:embed dev_0102_test.yaml
var testDataYaml_dev_0102 []byte

//go:embed dev_0102_test.json
var testDataJson_dev_0102 []byte

func TestRule_DEV_0102(t *testing.T) {
	var rule Rule_DEV_0102

	t.Run("test_json_rule_parsing", func(t *testing.T) {
		var jsonRule Rule_DEV_0102
		var yamlRule Rule_DEV_0102

		err := json.Unmarshal(testDataJson_dev_0102, &jsonRule)
		require.NoError(t, err)

		err = yaml.Unmarshal(testDataYaml_dev_0102, &yamlRule)
		require.NoError(t, err)

		require.Len(t, jsonRule.Restrict, 10)
		require.Len(t, yamlRule.Restrict, 10)
		require.EqualValues(t, jsonRule, yamlRule)
		rule = yamlRule
	})

	t.Run("test_rule_preparation", func(t *testing.T) {
		err := rule.Prepare()
		require.NoError(t, err)
		require.NotNil(t, rule.Exec)
	})

	t.Run("test_rule_execution", func(t *testing.T) {
		ok, violation := rule.Exec("21412")
		require.False(t, ok)
		require.Equal(t, "matched exact blacklisted statement", violation)

		ok, violation = rule.Exec("777")
		require.True(t, ok)
		require.Empty(t, violation)
	})
}
