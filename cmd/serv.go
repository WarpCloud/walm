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
	"walm/pkg/release/manager/helm"
	"os"
	"walm/pkg/k8s/elect"
	"walm/pkg/k8s/client"
	"encoding/json"
	"walm/pkg/k8s/informer/handlers"
	"walm/pkg/k8s/informer"
	"os/signal"
	"syscall"
	"walm/pkg/task"
	"context"
	"time"
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
	sc.initConfig()
	stopChan := make(chan struct{})
	informer.StartInformer(stopChan)
	task.GetDefaultTaskManager().StartWorker()
	startElect()
	initRestApi()

	server := &http.Server{Addr: fmt.Sprintf(":%d", setting.Config.HttpConfig.HTTPPort), Handler: restful.DefaultContainer}
	go func() {
		logrus.Error(server.ListenAndServe())
	}()
	logrus.Info("walm server started")
	// server.ListenAndServeTLS()

	//shut down gracefully
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	err := server.Shutdown(context.Background())
	if err != nil {
		logrus.Error(err.Error())
	}
	task.GetDefaultTaskManager().StopWorker()
	close(stopChan)
    logrus.Info("waiting for informer stopping")
	time.Sleep(2 * time.Second)
	logrus.Info("walm server stopped gracefully")
	return nil
}

func (sc *ServCmd)initConfig() {
	logrus.Infof("loading configuration from [%s]", sc.cfgFile)
	setting.InitConfig(sc.cfgFile)
	settingConfig, err := json.Marshal(setting.Config)
	if err != nil {
		logrus.Fatal("failed to marshal setting config")
	}
	logrus.Infof("finished loading configuration: %s", string(settingConfig))
}

func initRestApi() {
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
	restful.Add(router.InitSecretRouter())
	restful.Add(router.InitStorageClassRouter())
	restful.Add(router.InitPvcRouter())
	restful.Add(router.InitTenantRouter())
	restful.Add(router.InitProjectRouter())
	restful.Add(router.InitReleaseRouter())
	//restful.Add(router.InitReleaseV2Router())
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
	logrus.Infof("ready to serve on port %d", setting.Config.HttpConfig.HTTPPort)
}

func startElect() {
	lockIdentity := os.Getenv("Pod_Name")
	lockNamespace := os.Getenv("Pod_Namespace")
	if lockIdentity == "" || lockNamespace == "" {
		logrus.Fatal("Both env var Pod_Name and Pod_Namespace must not be empty")
	}

	onStartedLeadingFunc := func(stop <-chan struct{}) {
		logrus.Info("Succeed to elect leader")
		helm.GetDefaultHelmClient().StartResyncReleaseCaches(stop)
		handlers.StartHandlers(stop)
	}
	onNewLeaderFunc := func(identity string) {
		logrus.Infof("Now leader is changed to %s", identity)
	}
	onStoppedLeadingFunc := func() {
		logrus.Info("Stopped being a leader")
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
