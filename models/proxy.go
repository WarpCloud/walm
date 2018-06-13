package models

/*
var (
	appLock *sync.RWMutex
)
*/

func InsertAppInst(app AppInst) error {
	p := db.Begin()
	if err := p.Create(app).Error; err != nil {
		db.Rollback()
		return err
	}
	p.Commit()
	return nil
}

func InsertCluster(cluster *Cluster) error {
	p := db.Begin()
	if err := p.Create(*cluster).Error; err != nil {
		db.Rollback()
		return err
	}
	p.Commit()
	if err := p.Where("name = ?", cluster.Name).Find(cluster).Error; err != nil {
		return err
	}

	return nil
}

func UpdateAppInstByApp(app AppInst) error {
	p := db.Begin()
	if err := p.Update(app).Error; err != nil {
		db.Rollback()
		return err
	}
	p.Commit()
	return nil
}

func UpdateAppInst(name, value, actiontype string) error {
	appInst := &AppInst{Name: name}

	p := db.Begin()
	switch actiontype {
	case "status":
		if err := p.Model(appInst).Update("status", value).Error; err != nil {
			p.Rollback()
			return err
		}
	case "version":
		if err := p.Model(appInst).Update("vers", value).Error; err != nil {
			p.Rollback()
			return err
		}
	}

	p.Commit()
	return nil
}

func DeleteAppInst(name string) error {
	appInst := &AppInst{}

	p := db.Begin()
	if err := p.Where("name = ?", name).Find(appInst).Delete(&ClusterAppRefInst{AppInstId: appInst.AppInstId}).Delete(&AppInst{Name: name}).Error; err != nil {
		p.Rollback()
		return err
	}
	p.Commit()
	return nil
}

func DeleteCluster(name string) error {
	p := db.Begin()
	if err := p.Delete(Cluster{}, "name = ?", name).Error; err != nil {
		p.Rollback()
		return err
	}
	p.Commit()
	return nil
}

func GetReleasesOfCluster(clustername string) ([]string, error) {
	insts := []int{}
	var clusterid int

	if err := db.Where("name = ?", clustername).Select("cluster_id").Find(&clusterid).Select("app_inst_id").Find(insts, "cluster_id = ？", clusterid).Error; err != nil {
		return []string{}, err
	}

	namelist := []string{}

	if err := db.Where("app_inst_id in ?", insts).Select("name").Find(namelist).Error; err != nil {
		return namelist, err
	}

	return namelist, nil
}
