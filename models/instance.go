package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type AppInst struct {
	AppInstId     int    `gorm:"primary_key;auto_increment;column:app_inst_id"`
	Name          string `gorm:"column:name"`
	AppPkg        string `gorm:"column:app_pkg"`
	Vers          string `gorm:"column:vers"`
	ConfigTemp    string `gorm:"column:config_temp"`
	Status        string `gorm:"column:status"`
	InstallTime   int    `gorm:"column:install_time;type:int(11)"`
	InstalledTime int    `gorm:"column:installed_time;type:int(11)"`
	LastTime      int    `gorm:"column:last_time;type:int(11)"`
	ClusterId     int    `gorm:"column:cluster_id"`
}
type AppDepList struct {
	Id        int    `gorm:"primary_key;auto_increment;column:id"`
	AppPkg    string `gorm:"column:app_pkg"`
	AppListId int    `gorm:"column:app_list_id"`
	Vers      string `gorm:"column:vers"`
}

func (appInst *AppInst) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("InstallTime", time.Now().Unix())
	return nil
}

func (appInst *AppInst) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("LastTime", time.Now().Unix())
	return nil
}

type AppInstDepList struct {
	Id        int `gorm:"primary_key;auto_increment;column:id"`
	AppInstId int `gorm:"column:app_inst_id"`
	DepInstId int `gorm:"column:dep_inst_id"`
}
