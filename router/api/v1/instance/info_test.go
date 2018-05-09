package instance

import (
	"encoding/json"
	"fmt"
	"time"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
//func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&StatusSuite{})

type StatusSuite struct{}

func (stat *StatusSuite) Test_String(c *C) {
	datetime := time.Date(int(2018), time.March, int(30), int(12), int(12), int(12), int(0), time.FixedZone("UTC", int(8*60*60))).String()
	s := Info{
		Name:      "test",
		Revision:  "v1.0.0",
		Updated:   datetime,
		Status:    "deployed",
		Chart:     "testchart",
		Namespace: "namespace",
	}

	data, err := json.Marshal(s)
	c.Assert(err, IsNil, Commentf("json encode err: %s", err))

	dedata := &Info{}
	err = json.Unmarshal(data, dedata)
	c.Assert(err, IsNil, Commentf("json dncode err: %s", err))
	c.Assert(s, DeepEquals, *dedata, Commentf("change happend when encode and decode"))
	fmt.Println("test pass")
}

func (stat *StatusSuite) Test_StatusCode(c *C) {
	datetime := time.Date(int(2018), time.March, int(30), int(12), int(12), int(12), int(0), time.FixedZone("UTC", int(8*60*60))).String()
	//datetime := time.Now().Format(time.UnixDate)
	s := Info{
		Name:      "test",
		Revision:  "v1.0.0",
		Updated:   datetime,
		Status:    "deployed",
		Chart:     "testchart",
		Namespace: "namespace",
	}

	jdedata := &Info{}
	jsonb := `{"name":"test",
		"revision":"v1.0.0",
		"updated":"2018-03-30 12:12:12 +0800 UTC",
		"status":"deployed",
		"chart":"testchart",
		"namespace":"namespace"}`
	err := json.Unmarshal([]byte(jsonb), jdedata)
	c.Assert(err, IsNil, Commentf("json dncode err: %s", err))
	c.Assert(s, DeepEquals, *jdedata, Commentf("change happend when encode and decode actual json string"))

}
