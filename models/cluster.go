package models

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
