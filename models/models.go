package models

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	"walm/pkg/setting"
	. "walm/pkg/util/log"
)

var db *gorm.DB

func Init(conf *setting.Config) error {
	var err error
	if db, err = gorm.Open(conf.DbType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.DbUser,
		conf.DbPassword,
		conf.DbHost,
		conf.DbName)); err != nil {
		Log.Println(err)
		return err
	}

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		if conf.TablePrefix == "" {
			return defaultTableName
		}
		return conf.TablePrefix + defaultTableName
	}

	db.SetLogger(Log)
	db.LogMode(conf.Debug)

	db.SingularTable(true)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)

	return nil

}

func CloseDB() {
	defer db.Close()
}
