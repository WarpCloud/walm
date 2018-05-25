package models

type Cluster struct {
	ClusterId  int    `gorm:"primary_key;auto_increment;column:cluster_id"`
	Name       string `gorm:"column:name"`
	Namespace  string `gorm:"column:namespace"`
	ProdId     int    `gorm:"column:prod_id"`
	ConfigTemp string `gorm:"column:config_temp"`
}

func (Cluster) TableName() string {
	return "cluster"
}

type ClusterAppRefInst struct {
	ClusterAppRefInstId int `gorm:"primary_key;auto_increment;column:cluster_app_ref_inst_id"`
	ClusterId           int `gorm:"column:cluster_id"`
	AppInstId           int `gorm:"column:app_inst_id"`
}

func (ClusterAppRefInst) TableName() string {
	return "cluster_app_ref_inst"
}

type Product struct {
	ProdId     int    `gorm:"primary_key;auto_increment;column:prod_id"`
	Name       string `gorm:"column:name"`
	Note       string `gorm:"column:note"`
	AppListId  int    `gorm:"column:app_list_id"`
	Vers       string `gorm:"column:vers"`
	ConfigTemp string `gorm:"column:config_temp"`
}

func (Product) TableName() string {
	return "product"
}
