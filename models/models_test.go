// +build db

package models

import (
	"os"
	"strings"
	"testing"
	"walm/pkg/setting"

	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type modelSuite struct{}

var _ = check.TestingT(&modelSuite{})

func (ms *modelSuite) Test_Init(c *check.C) {
	dbhost := strings.Split(os.ExpandEnv("MYSQL_PORT_3306_TCP_ADDR"), "//")[1]
	conf := &setting.Config{
		Debug:      false,
		DbName:     "walm",
		DbPassword: "passwd",
		DbType:     "mysql",
		DbHost:     dbhost,
	}

	err := Init(conf)
	defer CloseDB()
	c.Assert(err, check.IsNil)
}
