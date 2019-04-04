package tenant

//func Test_CreateTenant(t *testing.T) {
//	//tenantParams := TenantParams{}
//	////tenantParams.TenantName = "walmbytest"
//	//err := CreateTenant(tenantParams)
//	//logrus.Errorf("%+v\n", err)
//}
//
//func Test_DeleteTenant(t *testing.T) {
//	err := DeleteTenant("walmbytest")
//	fmt.Printf("Test_DeleteTenant %+v\n", err)
//}
//
//func Test_GetTenant(t *testing.T) {
//	tenantInfo, err := GetTenant("walmbytest")
//	fmt.Printf("GeleteTenant %+v %+v\n", tenantInfo, err)
//}
//
//func TestMain(m *testing.M) {
//	gopath := os.Getenv("GOPATH")
//
//	setting.Config.KubeConfig = &setting.KubeConfig{
//		Config: gopath + "/src/walm/test/k8sconfig/kubeconfig",
//	}
//
//	logrus.Infof("loading configuration from [%s]", gopath + "/src/walm/walm.yaml")
//	setting.InitConfig(gopath + "/src/walm/walm.yaml")
//	settingConfig, err := json.Marshal(setting.Config)
//	if err != nil {
//		logrus.Fatal("failed to marshal setting config")
//	}
//	setting.Config.KubeConfig.Config = gopath + "/src/walm/test/k8sconfig/kubeconfig"
//	logrus.Infof("finished loading configuration: %s", string(settingConfig))
//
//	//kafka.InitKafkaClient(setting.Config.KafkaConfig)
//	//redis.InitRedisClient()
//	//job.InitWalmJobManager()
//	//informer.StartInformer()
//	//helm.InitHelm()
//	//project.InitProject()
//
//	os.Exit(m.Run())
//}
