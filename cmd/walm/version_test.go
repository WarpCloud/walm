package main

import (
	"testing"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type versionSuite struct{}

var _ = check.Suite(&versionSuite{})

func (vs *versionSuite) Test(c *check.C) {
	args := []string{"version"}
	cmd := newVersionCmd()
	err := cmd.RunE(cmd, args)
	c.Assert(err, check.IsNil)
}
