package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test_MergeValue(t *testing.T) {
	tests := []struct{
		destMap     map[string]interface{}
		srcMap      map[string]interface{}
		deleteKey   bool
		expectedMap map[string]interface{}
	} {
		{
			map[string]interface{}{
				"key1": "value1",
			},
			map[string]interface{}{
				"key1": "value2",
			},
			true,
			map[string]interface{}{
				"key1": "value2",
			},
		},
		{
			map[string]interface{}{
				"key1": "value1",
			},
			map[string]interface{}{
				"key2": "value2",
			},
			true,
			map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			map[string]interface{}{
				"key1": "value1",
			},
			map[string]interface{}{
				"key1": nil,
			},
			false,
			map[string]interface{}{
				"key1": nil,
			},
		},
		{
			map[string]interface{}{
				"key1": "value1",
			},
			map[string]interface{}{
				"key1": nil,
			},
			true,
			map[string]interface{}{
			},
		},
		{
			map[string]interface{}{
				"key1": "value1",
			},
			map[string]interface{}{
				"key2": nil,
			},
			true,
			map[string]interface{}{
				"key1": "value1",
			},
		},
		{
			map[string]interface{}{
				"key1": "value1",
			},
			map[string]interface{}{
				"key2": nil,
			},
			false,
			map[string]interface{}{
				"key1": "value1",
				"key2": nil,
			},
		},
		{
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"key2": "value2",
					"key1": "value1",
				},
			},
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"key1": "value2",
				},
			},
			true,
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"key2": "value2",
					"key1": "value2",
				},
			},
		},
		{
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"key1": "value1",
				},
			},
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"key2": "value2",
				},
			},
			true,
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"key2": "value2",
					"key1": "value1",
				},
			},
		},
		{
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"key2": "value2",
					"key1": "value1",
				},
			},
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"key1": nil,
				},
			},
			false,
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"key2": "value2",
					"key1": nil,
				},
			},
		},
		{
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"key2": "value2",
					"key1": "value1",
				},
			},
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"key1": nil,
				},
			},
			true,
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"key2": "value2",
				},
			},
		},
		{
			map[string]interface{}{
				"embededKey": map[string]interface{}{
					"embededKey2": map[string]interface{}{},
				},
			},
			map[string]interface{}{
				"embededKey": nil,
			},
			true,
			map[string]interface{}{
			},
		},
	}

	for _, test := range tests {
		MergeValues(test.destMap, test.srcMap, test.deleteKey)
		assert.Equal(t, test.expectedMap, test.destMap)
	}
}
