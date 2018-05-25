package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type AppInst struct {
	AppInstId     int    `gorm:"primary_key;auto_increment;column:app_inst_id"`
	Name          string `gorm:"column:name"`
	Namespace     string `gorm:"column:namespace"`
	AppPkg        string `gorm:"column:app_pkg"`
	Vers          string `gorm:"column:vers"`
	ConfigTemp    string `gorm:"column:config_temp"`
	Status        string `gorm:"column:status"`
	InstallTime   int64  `gorm:"column:install_time;type:int(11)"`
	InstalledTime int64  `gorm:"column:installed_time;type:int(11)"`
	LastTime      int64  `gorm:"column:last_time;type:int(11)"`
	ClusterId     int    `gorm:"column:cluster_id"`
}

func (AppInst) TableName() string {
	return "app_inst"
}

type AppDepList struct {
	Id        int    `gorm:"primary_key;auto_increment;column:id"`
	AppPkg    string `gorm:"column:app_pkg"`
	AppListId int    `gorm:"column:app_list_id"`
	Vers      string `gorm:"column:vers"`
}

func (AppDepList) TableName() string {
	return "app_dep_list"
}

type AppList struct {
	Id         int    `gorm:"primary_key;auto_increment;column:id"`
	AppPkg     string `gorm:"column:app_pkg"`
	AppListId  int    `gorm:"column:app_list_id"`
	Vers       string `gorm:"column:vers"`
	ConfigTemp string `gorm:"column:config_temp"`
}

func (AppList) TableName() string {
	return "app_list"
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

func (AppInstDepList) TableName() string {
	return "app_inst_dep_list"
}
