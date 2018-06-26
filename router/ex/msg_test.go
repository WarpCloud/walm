package ex

import (
	. "gopkg.in/check.v1"
)

type msgSuite struct {
}

var _ = Suite(&msgSuite{})

func (ms *msgSuite) Test_GetMsg(c *C) {
	str := GetMsg(SUCCESS)
	c.Assert(str, Equals, "ok")
	str = GetMsg(0)
	c.Assert(str, Equals, "Internal Server error")
}
