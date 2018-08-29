package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/emicklei/go-restful"
	"github.com/sirupsen/logrus"
	"github.com/go-openapi/spec"
	"github.com/emicklei/go-restful-openapi"

	"walm/router"
	"walm/router/middleware"
	"walm/pkg/setting"
	"walm/pkg/k8s/informer"
	"walm/pkg/release/manager/helm"
	"walm/pkg/release/manager/project"
	"walm/pkg/redis"
	"os"
	"walm/pkg/k8s/elect"
	"walm/pkg/k8s/client"
	"walm/pkg/job"
)

const servDesc = `
This command enable a WALM Web server.

$ walm serv 

Before to start serv ,you need to config the conf file 

The file is named conf.yaml

`

const DefaultElectionId = "walm-election-id"

type ServCmd struct {
	cfgFile string
}

func initService() error {
	redis.InitRedisClient()
	job.InitWalmJobManager()
	informer.InitInformer()
	helm.InitHelm()
	project.InitProject()

	return nil
}

func NewServCmd() *cobra.Command {
	inst := &ServCmd{}

	cmd := &cobra.Command{
		Use:   "serv",
		Short: "enable a Walm Web Server",
		Long:  servDesc,

		RunE: func(cmd *cobra.Command, args []string) error {
			return inst.run()
		},
	}
	cmd.PersistentFlags().StringVar(&inst.cfgFile, "config", "walm.yaml", "config file (default is walm.yaml)")

	return cmd
}

func (sc *ServCmd) run() error {
	logrus.Infof("loading configuration from [%s]", sc.cfgFile)
	setting.InitConfig(sc.cfgFile)
	logrus.Infof("finished loading configuration %+v", setting.Config)

	initService()
	initElector()
	// accept and respond in JSON unless told otherwise
	restful.DefaultRequestContentType(restful.MIME_JSON)
	restful.DefaultResponseContentType(restful.MIME_JSON)
	// gzip if accepted
	restful.DefaultContainer.EnableContentEncoding(true)
	// faster router
	restful.DefaultContainer.Router(restful.CurlyRouter{})
	restful.Filter(middleware.ServerStatsFilter)
	restful.Filter(middleware.RouteLogging)

	logrus.Infoln("Adding Route...")
	restful.Add(router.InitRootRouter())
	restful.Add(router.InitNodeRouter())
	restful.Add(router.InitTenantRouter())
	restful.Add(router.InitProjectRouter())
	restful.Add(router.InitReleaseRouter())
	restful.Add(router.InitPodRouter())
	restful.Add(router.InitChartRouter())
	logrus.Infoln("Add Route Success")

	config := restfulspec.Config{
		// You control what services are visible
		WebServices:                   restful.RegisteredWebServices(),
		APIPath:                       "/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	http.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("swagger-ui/dist"))))
	http.Handle("/swagger/", http.RedirectHandler("/swagger-ui/?url=/apidocs.json", http.StatusFound))

	logrus.Infoln("ready to serve on")
	server := &http.Server{Addr: fmt.Sprintf(":%d", setting.Config.HttpConfig.HTTPPort), Handler: restful.DefaultContainer}
	logrus.Fatalln(server.ListenAndServe())
	// server.ListenAndServeTLS()

	return nil
}

func initElector() {
	lockIdentity := os.Getenv("Pod_Name")
	lockNamespace := os.Getenv("Pod_Namespace")
	if lockIdentity == "" || lockNamespace == "" {
		logrus.Fatal("Both env var Pod_Name and Pod_Namespace must not be empty")
	}

	onStartedLeadingFunc := func(stop <-chan struct{}) {
		logrus.Info("Succeed to elect leader")
		helm.GetDefaultHelmClient().StartResyncReleaseCaches(stop)
		job.GetDefaultWalmJobManager().Start(stop)
	}
	onNewLeaderFunc := func(identity string) {
		logrus.Infof("Now leader is changed to %s", identity)
	}
	onStoppedLeadingFunc := func() {
		logrus.Info("Stopped being a leader")
		job.GetDefaultWalmJobManager().Stop()
	}

	config := &elect.ElectorConfig{
		LockNamespace:        lockNamespace,
		LockIdentity:         lockIdentity,
		ElectionId:           DefaultElectionId,
		Client:               client.GetDefaultClient(),
		OnStartedLeadingFunc: onStartedLeadingFunc,
		OnNewLeaderFunc:      onNewLeaderFunc,
		OnStoppedLeadingFunc: onStoppedLeadingFunc,
	}

	elector, err := elect.NewElector(config)
	if err != nil {
		logrus.Fatal("create leader elector failed")
	}
	logrus.Info("Start to elect leader")
	go elector.Run()
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Walm",
			Description: "Walm Web Server",
			Version:     "0.0.1",
		},
	}
}
