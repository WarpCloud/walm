// +build db

package main

import (
	"os"
	"strings"

	"gopkg.in/check.v1"
)

type migrateSuite struct{}

var _ = check.Suite(&migrateSuite{})

func (ms *migrateSuite) Test(c *check.C) {
	args := []string{"migrate"}

	dbhost := strings.Split(os.ExpandEnv("MYSQL_PORT_3306_TCP_ADDR"), "//")[1]
	conf.DbUser = "root"
	conf.DbPassword = "passwd"
	conf.DbHost = dbhost
	conf.DbType = "mysql"
	conf.DbName = "walm"

	cmd := newMigrateCmd()
	err := cmd.RunE(cmd, args)
	c.Assert(err, check.IsNil)
}
