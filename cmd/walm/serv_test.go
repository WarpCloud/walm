// +build db,unit

package main

import (
	"os"
	"strings"

	"gopkg.in/check.v1"
)

type servSuite struct{}

var _ = check.Suite(&servSuite{})

func (ss *servSuite) Test(c *check.C) {
	dbhost := strings.Split(os.ExpandEnv("MYSQL_PORT_3306_TCP_ADDR"), "//")[1]
	conf.DbUser = "root"
	conf.DbPassword = "passwd"
	conf.DbHost = dbhost
	conf.DbType = "mysql"
	conf.DbName = "walm"
	conf.Debug = true

	cmd := newServCmd()
	err := cmd.RunE(cmd, args)

	c.Assert(err, check.IsNil)
}
