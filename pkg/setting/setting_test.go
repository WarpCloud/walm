package setting

import (
	"os"

	"gopkg.in/check.v1"
)

type settingSuit struct{}

var _ = check.Suite(&settingSuit{})

func (ss *settingSuit) Test(c *check.C) {
	c.Assert(Config.Http.HTTPPort, check.Equals, 9000)
	c.Assert(Config.Home, check.Equals, os.Getenv("HOME")+".walm")
}
