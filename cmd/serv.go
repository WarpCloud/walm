package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/emicklei/go-restful"
	"github.com/sirupsen/logrus"
	"github.com/go-openapi/spec"
	"github.com/emicklei/go-restful-openapi"

	"WarpCloud/walm/router"
	"WarpCloud/walm/router/middleware"
	"WarpCloud/walm/pkg/setting"
	"WarpCloud/walm/pkg/release/manager/helm"
	"os"
	"WarpCloud/walm/pkg/k8s/elect"
	"WarpCloud/walm/pkg/k8s/client"
	"encoding/json"
	"WarpCloud/walm/pkg/k8s/informer/handlers"
	"WarpCloud/walm/pkg/k8s/informer"
	"os/signal"
	"syscall"
	"WarpCloud/walm/pkg/task"
	"context"
	"time"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	transwarpscheme "transwarp/release-config/pkg/client/clientset/versioned/scheme"
	"github.com/x-cray/logrus-prefixed-formatter"
	_ "net/http/pprof"
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
	sig := make(chan os.Signal, 1)

	logrus.SetFormatter(&prefixed.TextFormatter{})
	sc.initConfig()
	initLogLevel()
	stopChan := make(chan struct{})
	transwarpscheme.AddToScheme(clientsetscheme.Scheme)

	informer.StartInformer(stopChan)
	task.GetDefaultTaskManager().StartWorker()
	startElect(stopChan, sig)
	initRestApi()

	if setting.Config.Debug {
		go func() {
			logrus.Info("supporting pprof on port 6060...")
			logrus.Error(http.ListenAndServe(":6060", nil))
		}()
	}

	server := &http.Server{Addr: fmt.Sprintf(":%d", setting.Config.HttpConfig.HTTPPort), Handler: restful.DefaultContainer}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logrus.Error(err.Error())
			sig <- syscall.SIGINT
		}
	}()

	//shut down gracefully
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

func initLogLevel() {
	if setting.Config.LogConfig != nil {
		level, err := logrus.ParseLevel(setting.Config.LogConfig.Level)
		if err != nil {
			logrus.Warnf("failed to parse log level %s : %s", setting.Config.LogConfig.Level, err.Error())
		} else {
			logrus.SetLevel(level)
			logrus.Infof("log level is set to %s", setting.Config.LogConfig.Level)
		}
	}
}

func (sc *ServCmd) initConfig() {
	logrus.Infof("loading configuration from [%s]", sc.cfgFile)
	setting.InitConfig(sc.cfgFile)
	settingConfig, err := json.MarshalIndent(setting.Config, "", "  ")
	if err != nil {
		logrus.Fatal("failed to marshal setting config")
	}
	logrus.Infof("finished loading configuration:\n%s", string(settingConfig))
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

func startElect(stopCh <-chan struct{}, sig chan os.Signal) {
	// TODO investigate it (from scheduler)
	ctx, cancel := context.WithCancel(context.TODO())
	go func() {
		select {
		case <-stopCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	lockIdentity := os.Getenv("Pod_Name")
	lockNamespace := os.Getenv("Pod_Namespace")
	if lockIdentity == "" || lockNamespace == "" {
		logrus.Fatal("Both env var Pod_Name and Pod_Namespace must not be empty")
	}

	onStartedLeadingFunc := func(context context.Context) {
		logrus.Info("Succeed to elect leader")
		helm.GetDefaultHelmClient().StartResyncReleaseCaches(context.Done())
		handlers.StartHandlers(context.Done())
	}
	onNewLeaderFunc := func(identity string) {
		logrus.Infof("Now leader is changed to %s", identity)
	}
	onStoppedLeadingFunc := func() {
		logrus.Info("Stopped being a leader")
		sig <- syscall.SIGINT
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
	go elector.Run(ctx)
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
