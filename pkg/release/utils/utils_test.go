package utils

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"errors"
)

func Test_ConfigValuesDiff(t *testing.T) {
	tests := []struct {
		configValue1 map[string]interface{}
		configValue2 map[string]interface{}
		diff         bool
	}{
		{
			configValue2: map[string]interface{}{},
			diff:         false,
		},
		{
			configValue2: map[string]interface{}{"test": "true"},
			diff:         true,
		},
	}

	for _, test := range tests {
		diff := ConfigValuesDiff(test.configValue1, test.configValue2)
		assert.Equal(t, test.diff, diff)
	}
}

func Test_ParseDependedRelease(t *testing.T) {
	tests := []struct {
		dependingReleaseNamespace string
		dependedRelease           string
		namespace                 string
		name                      string
		err                       error
	}{
		{
			dependingReleaseNamespace: "testns",
			dependedRelease: "testnm",
			namespace: "testns",
			name: "testnm",
			err: nil,
		},
		{
			dependingReleaseNamespace: "testns",
			dependedRelease: "testns1/testnm",
			namespace: "testns1",
			name: "testnm",
			err: nil,
		},
		{
			dependingReleaseNamespace: "testns",
			dependedRelease: "testns1/testnm/error",
			namespace: "",
			name: "",
			err: errors.New(""),
		},
	}

	for _, test := range tests {
		namespace, name, err := ParseDependedRelease(test.dependingReleaseNamespace, test.dependedRelease)
		assert.IsType(t, test.err, err)
		assert.Equal(t, test.namespace, namespace)
		assert.Equal(t, test.name, name)
	}
}
