package transwarpjsonnet

import (
	"fmt"
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
		"weblogic/template-jsonnet/weblogic.jsonnet": testFuncFileStr,
	}
	configValues := map[string]interface{}{
		"key1": "value1",
	}

	resourceStr, err := renderMainJsonnetFile(templateFiles, configValues)
	if err != nil {

	}
	fmt.Printf("%s", resourceStr)
	assert.Equal(t, resourceStr, renderedFileStr)
}
