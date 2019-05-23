package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var r = Rules{
	"rules": map[string]interface{}{
		"profiles": map[string]interface{}{
			"$user_id": map[string]interface{}{
				"name": map[string]interface{}{
					".validate": "newData.isString()",
				},
				"email": map[string]interface{}{
					".validate": "newData.isString()",
				},
				"uid": map[string]interface{}{
					".validate": "newData.isString()",
				},
				"$other": map[string]interface{}{
					".validate": "false",
				},
				".read":  "auth != null && auth.uid == $user_id",
				".write": "auth != null && auth.uid == $user_id",
			},
			".indexOn": []string{"email"},
		},
	},
}

func TestRulesChild(t *testing.T) {
	var nilRule Rules
	assert.Nil(t, nilRule.Child("missing"))
	assert.Nil(t, r.Child("missing"))
	assert.EqualValues(t, Rules{
		"profiles": map[string]interface{}{
			"$user_id": map[string]interface{}{
				"name": map[string]interface{}{
					".validate": "newData.isString()",
				},
				"email": map[string]interface{}{
					".validate": "newData.isString()",
				},
				"uid": map[string]interface{}{
					".validate": "newData.isString()",
				},
				"$other": map[string]interface{}{
					".validate": "false",
				},
				".read":  "auth != null && auth.uid == $user_id",
				".write": "auth != null && auth.uid == $user_id",
			},
			".indexOn": []string{"email"},
		},
	}, r.Child("rules"))
	assert.EqualValues(t, Rules{
		"$user_id": map[string]interface{}{
			"name": map[string]interface{}{
				".validate": "newData.isString()",
			},
			"email": map[string]interface{}{
				".validate": "newData.isString()",
			},
			"uid": map[string]interface{}{
				".validate": "newData.isString()",
			},
			"$other": map[string]interface{}{
				".validate": "false",
			},
			".read":  "auth != null && auth.uid == $user_id",
			".write": "auth != null && auth.uid == $user_id",
		},
		".indexOn": []string{"email"},
	}, r.Child("rules").Child("profiles"))
	assert.EqualValues(t, Rules{
		"name": map[string]interface{}{
			".validate": "newData.isString()",
		},
		"email": map[string]interface{}{
			".validate": "newData.isString()",
		},
		"uid": map[string]interface{}{
			".validate": "newData.isString()",
		},
		"$other": map[string]interface{}{
			".validate": "false",
		},
		".read":  "auth != null && auth.uid == $user_id",
		".write": "auth != null && auth.uid == $user_id",
	}, r.Child("rules").Child("profiles").Child("user1"))
	assert.EqualValues(t, Rules{
		".validate": "newData.isString()",
	}, r.Child("rules").Child("profiles").Child("user1").Child("name"))
	assert.EqualValues(t, Rules{
		".validate": "newData.isString()",
	}, r.Child("rules/profiles/user1/name"))
	assert.Nil(t, r.Child("rules").Child("profiles").Child("user1").Child("name").Child("missing"))
	assert.Nil(t, r.Child("rules/profiles/user1/name/missing"))
}

func TestIndexes(t *testing.T) {
	var nilRule Rules
	assert.Nil(t, nilRule.Indexes())
	assert.Nil(t, r.Indexes())
	assert.Nil(t, r.Child("missing").Indexes())
	assert.Nil(t, r.Child("rules").Indexes())
	assert.EqualValues(t, []string{"email"}, r.Child("/rules/profiles").Indexes())
	assert.Nil(t, r.Child("/rules/profiles/user1").Indexes())
}

func TestContainsKey(t *testing.T) {
	var nilRule Rules
	assert.False(t, nilRule.ContainsKey("missing"))
	assert.False(t, r.Child("/rules/profiles/user1").ContainsKey("phone"))
	assert.True(t, r.Child("/rules/profiles/user1").ContainsKey("name"))
	assert.True(t, r.Child("/rules/profiles/user1").ContainsKey("$other"))
}

func TestContainsVariableKey(t *testing.T) {
	var nilRule Rules
	assert.Empty(t, nilRule.VariableKey())
	assert.Empty(t, r.Child("/rules").VariableKey())
	assert.Equal(t, "$user_id", r.Child("/rules/profiles").VariableKey())
	assert.Empty(t, r.Child("/rules/profiles/user1").VariableKey())
}
