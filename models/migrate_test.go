// +build db

package models

import (
	"walm/pkg/setting"

	"gopkg.in/check.v1"
)

type atuoMigrateSuite struct{}

var _ = check.Suite(&atuoMigrateSuite{})

func (ams *atuoMigrateSuite) Test_AutoMigrate(c *check.C) {
	conf := &setting.Config{
		DbUser:     "root",
		DbPassword: "passwd",
		DbHost:     "",
		DbType:     "mysql",
		DbName:     "walm",
	}
	err := AutoMigrate(conf)
	c.Assert(err, check.IsNil)
}
