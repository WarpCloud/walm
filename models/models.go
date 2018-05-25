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
	var db *gorm.DB
	var err error
	if db, err = gorm.Open(conf.DbType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
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

func AutoMigrate() {
	//auto migrate cluster
	db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&Cluster{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&ClusterAppRefInst{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&Product{})
	//auto migrate instance
	db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&Product{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&Product{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&Product{})
	//auto migrate event
	db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&Product{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&Product{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&Product{})
}

func CreateProductTable() bool {
	db.DropTableIfExists("product",
		"app_list",
		"app_dep_list",
		"cluster",
		"cluster_app_ref_inst",
		"app_inst",
		"app_inst_dep_list")

	db.Set("gorm:table_options", "ENGINE=InnoDB").Table("product").CreateTable(&Product{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Table("app_list").CreateTable(&AppList{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Table("app_dep_list").CreateTable(&AppDepList{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Table("cluster").CreateTable(&Cluster{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Table("cluster_app_ref_inst").CreateTable(&ClusterAppRefInst{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Table("app_inst").CreateTable(&AppInst{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Table("app_inst_dep_list").CreateTable(&AppInstDepList{})

	if len(db.GetErrors()) > 0 {
		return false
	}

	return true
}

func CreateEventTable() bool {
	db.DropTableIfExists("event",
		"evnet_action",
		"event_deal_inst",
		"event_deal_rule")

	db.Set("gorm:table_options", "ENGINE=InnoDB").Table("event").CreateTable(&Event{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Table("evnet_action").CreateTable(&EvnetAction{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Table("event_deal_inst").CreateTable(&EventDealInst{})
	db.Set("gorm:table_options", "ENGINE=InnoDB").Table("event_deal_rule").CreateTable(&EventDealRule{})

	if len(db.GetErrors()) > 0 {
		return false
	}

	return true
}
