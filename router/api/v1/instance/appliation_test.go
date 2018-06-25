package instance

import (
	"encoding/json"

	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type ApplicationSuite struct {
}

var _ = Suite(&ApplicationSuite{})

func (app *ApplicationSuite) TestApplicationString(c *C) {
	//datetime := time.Now().Format(time.UnixDate)
	valueStr := "key1=value1,key2=value2"
	apps := Application{
		Name:      "test",
		Namespace: "namespace",
		Version:   "v1.0.1",
		Value:     valueStr,
		Install:   true,
	}

	jdedata := &Application{}
	jsonb := `{"name":"test","namespace":"namespace","version":"v1.0.1","value":"key1=value1,key2=value2","install":true}`
	err := json.Unmarshal([]byte(jsonb), jdedata)
	c.Assert(err, IsNil, Commentf("json dncode err: %s", err))
	c.Assert(apps, DeepEquals, *jdedata, Commentf("change happend when encode and decode actual json string"))

}
