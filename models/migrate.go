package models

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	"walm/pkg/setting"
)

func createNoneExistDB(conf *setting.Config) error {
	dbconf := fmt.Sprintf("%s:%s@%s/", conf.DbUser, conf.DbPassword, conf.DbHost)
	if datebase, err := sql.Open(conf.DbType, dbconf); err != nil {
		return err
	} else {
		dbstr := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci", conf.DbName)
		if _, err := datebase.Exec(dbstr); err != nil {
			return err
		}
		datebase.Close()
	}
	return nil
}

func AutoMigrate(conf *setting.Config) error {
	if err := createNoneExistDB(conf); err != nil {
		return errors.New("create database failed!")
	}
	var err error
	if db, err = gorm.Open(conf.DbType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		conf.DbUser,
		conf.DbPassword,
		conf.DbHost,
		conf.DbName)); err != nil {
		return errors.New("connect to database failed!")
	} else {
		//auto migrate cluster
		db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&Cluster{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&ClusterAppRefInst{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&Product{})
		//auto migrate instance
		db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&AppInst{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&AppDepList{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&AppList{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&AppInstDepList{})
		//auto migrate event
		db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&Event{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&EventDealRule{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&EventDealInst{})
		db.Set("gorm:table_options", "ENGINE=InnoDB").Set("gorm:table_options", "CHARSET=utf8").AutoMigrate(&EvnetAction{})

	}
	return nil
}

/*
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
*/
