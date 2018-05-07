package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Product struct {
	ProdId     int    `gorm:"primary_key;auto_increment;column:prod_id"`
	Name       string `gorm:"column:name"`
	Note       string `gorm:"column:note"`
	AppListId  int    `gorm:"column:app_list_id"`
	Vers       string `gorm:"column:vers"`
	ConfigTemp string `gorm:"column:config_temp"`
}

type AppList struct {
	Id         int    `gorm:"primary_key;auto_increment;column:id"`
	AppPkg     string `gorm:"column:app_pkg"`
	AppListId  int    `gorm:"column:app_list_id"`
	Vers       string `gorm:"column:vers"`
	ConfigTemp string `gorm:"column:config_temp"`
}

type AppDepList struct {
	Id        int    `gorm:"primary_key;auto_increment;column:id"`
	AppPkg    string `gorm:"column:app_pkg"`
	AppListId int    `gorm:"column:app_list_id"`
	Vers      string `gorm:"column:vers"`
}

type Cluster struct {
	ClusterId  int    `gorm:"primary_key;auto_increment;column:cluster_id"`
	ProdId     int    `gorm:"column:prod_id"`
	ConfigTemp string `gorm:"column:config_temp"`
}

type ClusterAppRefInst struct {
	ClusterAppRefInstId int `gorm:"primary_key;auto_increment;column:cluster_app_ref_inst_id"`
	ClusterId           int `gorm:"column:cluster_id"`
	AppInstId           int `gorm:"column:app_inst_id"`
}

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
