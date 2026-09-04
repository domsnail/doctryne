package rules

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

//go:embed email_0101_test.yaml
var testDataYaml_email_0101 []byte

//go:embed email_0101_test.json
var testDataJson_email_0101 []byte

func TestRule_EMAIL_0101(t *testing.T) {
	var rule Rule_EMAIL_0101

	t.Run("test_json_rule_parsing", func(t *testing.T) {
		var jsonRule Rule_EMAIL_0101
		var yamlRule Rule_EMAIL_0101

		err := json.Unmarshal(testDataJson_email_0101, &jsonRule)
		require.NoError(t, err)

		err = yaml.Unmarshal(testDataYaml_email_0101, &yamlRule)
		require.NoError(t, err)

		require.Len(t, jsonRule.Restrict, 200)
		require.Len(t, jsonRule.Allow, 100)
		require.Len(t, yamlRule.Restrict, 200)
		require.Len(t, yamlRule.Allow, 100)
		require.EqualValues(t, jsonRule, yamlRule)
		rule = yamlRule
	})

	t.Run("test_rule_preparation", func(t *testing.T) {
		err := rule.Prepare()
		require.NoError(t, err)

		require.NotNil(t, "test-0101", rule.Exec)
	})

	t.Run("test_rule_execution", func(t *testing.T) {
		ok, violation := rule.Exec("ibraheem.mcdaniel@yahoo.com")
		require.True(t, ok)
		require.Empty(t, violation)

		ok, violation = rule.Exec("kailin.petty@aol.com")
		require.False(t, ok)
		require.Equal(t, "matched exact blacklisted statement", violation)

		ok, violation = rule.Exec("banned.petty@test.eu")
		require.False(t, ok)
		require.Equal(t, "matched blacklisted contains statement", violation)

		ok, violation = rule.Exec("banned_exception@yandex.ru")
		require.True(t, ok)
		require.Empty(t, violation)

		ok, violation = rule.Exec("test@regexp.com")
		require.False(t, ok)
		require.Equal(t, "matched blacklisted regexp", violation)
	})
}
