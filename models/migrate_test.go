// +build db

package models

import (
	"os"
	"strings"
	"walm/pkg/setting"

	"gopkg.in/check.v1"
)

type atuoMigrateSuite struct{}

var _ = check.Suite(&atuoMigrateSuite{})

func (ams *atuoMigrateSuite) Test_AutoMigrate(c *check.C) {
	dbhost := strings.Split(os.ExpandEnv("MYSQL_PORT_3306_TCP_ADDR"), "//")[1]
	conf := &setting.Config{
		DbUser:     "root",
		DbPassword: "passwd",
		DbHost:     dbhost,
		DbType:     "mysql",
		DbName:     "walm",
	}
	err := AutoMigrate(conf)
	c.Assert(err, check.IsNil)
}
