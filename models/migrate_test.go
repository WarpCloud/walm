// +build db

package models

import (
	"testing"
	"walm/pkg/setting"
)

func TestAutoMigrate(t *testing.T) {
	conf := &setting.Config{
		DbUser:     "root",
		DbPassword: "passwd",
		DbHost:     "",
		DbType:     "mysql",
		DbName:     "walm",
	}
	if err := AutoMigrate(conf); err != nil {
		t.Fatal(err)
	}
}
