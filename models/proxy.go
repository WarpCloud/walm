package models

import (
	"errors"
)

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

/*
var (
	appLock *sync.RWMutex
)
*/

func DeleteAppInst(name string) error {
	appInst := &AppInst{}

	/*
		* lock or transaction ??? *

		appLock.Lock()
		defer appLock.Unlock()
		db.Where("name = ?", name).Find(appInst)
		db.Delete(&ClusterAppRefInst{AppInstId: appInst.AppInstId})
		db.Delete(&AppInst{Name: name})
	*/

	db.Begin().Where("name = ?", name).Find(appInst).Delete(&ClusterAppRefInst{AppInstId: appInst.AppInstId}).Delete(&AppInst{Name: name})

	if len(db.GetErrors()) > 0 {
		db.Rollback()
		return errors.New("delete instance error!")
	}
	db.Commit()
	return nil
}
