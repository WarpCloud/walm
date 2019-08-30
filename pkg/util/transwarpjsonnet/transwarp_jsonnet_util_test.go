package transwarpjsonnet

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testMainFileStr = `
local weblogic = import './weblogic.jsonnet';

function(config={})
  {
    "test": weblogic.func()
  }
`

var testFuncFileStr = `
{
  func():: {
    'test': 'abcd'
  }
}
`

var renderedFileStr = `{
   "test": {
      "test": "abcd"
   }
}
`

func Test_renderMainJsonnetFile(t *testing.T) {
	templateFiles := map[string]string{
		"weblogic/template-jsonnet/weblogic-main.jsonnet": testMainFileStr,
		"weblogic/template-jsonnet/weblogic.jsonnet":      testFuncFileStr,
	}
	configValues := map[string]interface{}{
		"key1": "value1",
	}

	resourceStr, err := renderMainJsonnetFile(templateFiles, configValues)
	if err != nil {
		panic(err)
	}
	var tmpObj interface{}
	_ = json.Unmarshal([]byte(resourceStr), &tmpObj)
	resourceStrBytes, _ := json.Marshal(tmpObj)
	_ = json.Unmarshal([]byte(renderedFileStr), &tmpObj)
	renderedFileStrBytes, _ := json.Marshal(tmpObj)
	assert.Equal(t, resourceStrBytes, renderedFileStrBytes)
}

func Test_gotmplRender(t *testing.T) {
	type testVal struct {
		Name string
	}
	val1 := testVal{
		Name: "test",
	}
	_, err := gotmplRender("{{ .Name }}", val1, "string")
	assert.Nil(t, err, "gotmplRender error")
}

func Test_getMainJsonnetFile(t *testing.T) {
	templateFiles := map[string]string{
		"test-main.jsonnet": "content",
	}
	fileName, err := getMainJsonnetFile(templateFiles)
	assert.Nil(t, err, "getMainJsonnetFile error")
	assert.Equal(t, fileName, "test-main.jsonnet")

	template2Files := map[string]string{
		"test.jsonnet": "content",
	}
	_, err = getMainJsonnetFile(template2Files)
	assert.NotNil(t, err)
}

func Test_buildKubeResourcesByJsonStr(t *testing.T) {
	_, err := buildKubeResourcesByJsonStr("")
	assert.NotNil(t, err)
}