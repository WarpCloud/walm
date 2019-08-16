package cmd

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/emicklei/go-restful"
	"github.com/sirupsen/logrus"
	"github.com/go-openapi/spec"
	"github.com/emicklei/go-restful-openapi"

	"WarpCloud/walm/pkg/setting"
	"os"
	"WarpCloud/walm/pkg/k8s/elect"
	"WarpCloud/walm/pkg/k8s/client"
	"encoding/json"
	"os/signal"
	"syscall"
	"context"
	"time"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	transwarpscheme "transwarp/release-config/pkg/client/clientset/versioned/scheme"
	"github.com/x-cray/logrus-prefixed-formatter"
	_ "net/http/pprof"
	cacheInformer "WarpCloud/walm/pkg/k8s/cache/informer"
	"WarpCloud/walm/pkg/task/machinery"
	"errors"
	"WarpCloud/walm/pkg/redis/impl"
	helmImpl "WarpCloud/walm/pkg/helm/impl"
	k8sHelm "WarpCloud/walm/pkg/k8s/client/helm"
	"WarpCloud/walm/pkg/sync"
	releaseconfig "WarpCloud/walm/pkg/release/config"
	releaseusecase "WarpCloud/walm/pkg/release/usecase/helm"
	"WarpCloud/walm/pkg/k8s/operator"
	releasecache "WarpCloud/walm/pkg/release/cache/redis"
	kafkaimpl "WarpCloud/walm/pkg/kafka/impl"
	projectusecase "WarpCloud/walm/pkg/project/usecase"
	projectcache "WarpCloud/walm/pkg/project/cache/redis"
	httpModel "WarpCloud/walm/pkg/models/http"
	nodehttp "WarpCloud/walm/pkg/node/delivery/http"
	secrethttp "WarpCloud/walm/pkg/secret/delivery/http"
	storageclasshttp "WarpCloud/walm/pkg/storageclass/delivery/http"
	pvchttp "WarpCloud/walm/pkg/pvc/delivery/http"
	tenanthttp "WarpCloud/walm/pkg/tenant/delivery/http"
	tenantusecase "WarpCloud/walm/pkg/tenant/usecase"
	projecthttp "WarpCloud/walm/pkg/project/delivery/http"
	releasehttp "WarpCloud/walm/pkg/release/delivery/http"
	podhttp "WarpCloud/walm/pkg/pod/delivery/http"
	"github.com/thoas/stats"
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
	lockIdentity := os.Getenv("Pod_Name")
	lockNamespace := os.Getenv("Pod_Namespace")
	if lockIdentity == "" || lockNamespace == "" {
		err := errors.New("both env var Pod_Name and Pod_Namespace must not be empty")
		logrus.Error(err.Error())
		return err
	}

	sig := make(chan os.Signal, 1)

	logrus.SetFormatter(&prefixed.TextFormatter{})
	sc.initConfig()
	config := setting.Config
	initLogLevel()
	stopChan := make(chan struct{})
	transwarpscheme.AddToScheme(clientsetscheme.Scheme)

	kubeConfig := ""
	if config.KubeConfig != nil {
		kubeConfig = config.KubeConfig.Config
	}
	k8sClient, err := client.NewClient("", kubeConfig)
	if err != nil {
		logrus.Errorf("failed to create k8s client : %s", err.Error())
		return err
	}
	k8sReleaseConfigClient, err := client.NewReleaseConfigClient("", kubeConfig)
	if err != nil {
		logrus.Errorf("failed to create k8s release config client : %s", err.Error())
		return err
	}
	k8sCache := cacheInformer.NewInformer(k8sClient, k8sReleaseConfigClient, 0, stopChan)

	if config.TaskConfig == nil {
		err = errors.New("task config can not be empty")
		logrus.Error(err.Error())
		return err
	}
	task, err := machinery.NewTask(config.TaskConfig)
	if err != nil {
		logrus.Errorf("failed to create task manager %s", err.Error())
		return err
	}

	registryClient := helmImpl.NewRegistryClient(config.ChartImageConfig)
	kubeClients := k8sHelm.NewHelmKubeClient(kubeConfig)
	helm, err := helmImpl.NewHelm(config.RepoList, registryClient, k8sCache, kubeClients)
	if err != nil {
		logrus.Errorf("failed to create helm manager: %s", err.Error())
		return err
	}
	k8sOperator := operator.NewOperator(k8sClient, k8sCache, kubeClients)
	if config.RedisConfig == nil {
		err = errors.New("redis config can not be empty")
		logrus.Error(err.Error())
		return err
	}
	redisClient := impl.NewRedisClient(config.RedisConfig)
	redis := impl.NewRedis(redisClient)
	releaseCache := releasecache.NewCache(redis)
	releaseUseCase, err := releaseusecase.NewHelm(releaseCache, helm, k8sCache, k8sOperator, task)
	if err != nil {
		logrus.Errorf("failed to new release use case : %s", err.Error())
		return err
	}
	projectCache := projectcache.NewProjectCache(redis)
	projectUseCase, err := projectusecase.NewProject(projectCache, task, releaseUseCase, helm)
	if err != nil {
		logrus.Errorf("failed to new project use case : %s", err.Error())
		return err
	}

	ctx, cancel := context.WithCancel(context.TODO())
	go func() {
		select {
		case <-stopChan:
			cancel()
		case <-ctx.Done():
		}
	}()

	syncManager := sync.NewSync(redisClient, helm, k8sCache, task, "", "", "")
	kafka, err := kafkaimpl.NewKafka(config.KafkaConfig)
	if err != nil {
		logrus.Errorf("failed to create kafka manager: %s", err.Error())
		return err
	}
	releaseConfigController := releaseconfig.NewReleaseConfigController(k8sCache, releaseUseCase, kafka, 0)
	onStartedLeadingFunc := func(context context.Context) {
		logrus.Info("Succeed to elect leader")
		syncManager.Start(context.Done())
		releaseConfigController.Start(context.Done())
	}
	onNewLeaderFunc := func(identity string) {
		logrus.Infof("Now leader is changed to %s", identity)
	}
	onStoppedLeadingFunc := func() {
		logrus.Info("Stopped being a leader")
		sig <- syscall.SIGINT
	}

	electorConfig := &elect.ElectorConfig{
		LockNamespace:        lockNamespace,
		LockIdentity:         lockIdentity,
		ElectionId:           DefaultElectionId,
		Client:               k8sClient,
		OnStartedLeadingFunc: onStartedLeadingFunc,
		OnNewLeaderFunc:      onNewLeaderFunc,
		OnStoppedLeadingFunc: onStoppedLeadingFunc,
	}

	elector, err := elect.NewElector(electorConfig)
	if err != nil {
		logrus.Errorf("create leader elector failed")
		return err
	}
	logrus.Info("Start to elect leader")
	go elector.Run(ctx)

	restful.DefaultRequestContentType(restful.MIME_JSON)
	restful.DefaultResponseContentType(restful.MIME_JSON)
	// gzip if accepted
	restful.DefaultContainer.EnableContentEncoding(true)
	// faster router
	restful.DefaultContainer.Router(restful.CurlyRouter{})
	restful.Filter(ServerStatsFilter)
	restful.Filter(RouteLogging)
	logrus.Infoln("Adding Route...")

	restful.Add(InitRootRouter())
	restful.Add(nodehttp.RegisterNodeHandler(k8sCache, k8sOperator))
	restful.Add(secrethttp.RegisterSecretHandler(k8sCache, k8sOperator))
	restful.Add(storageclasshttp.RegisterStorageClassHandler(k8sCache))
	restful.Add(pvchttp.RegisterPvcHandler(k8sCache, k8sOperator))
	tenantUseCase := tenantusecase.NewTenant(k8sCache, k8sOperator, releaseUseCase)
	restful.Add(tenanthttp.RegisterTenantHandler(tenantUseCase))
	restful.Add(projecthttp.RegisterProjectHandler(projecthttp.NewProjectHandler(projectUseCase)))
	restful.Add(releasehttp.RegisterReleaseHandler(releasehttp.NewReleaseHandler(releaseUseCase)))
	restful.Add(podhttp.RegisterPodHandler(k8sCache, k8sOperator))
	restful.Add(releasehttp.RegisterChartHandler(helm))
	logrus.Infoln("Add Route Success")
	restConfig := restfulspec.Config{
		// You control what services are visible
		WebServices:                   restful.RegisteredWebServices(),
		APIPath:                       "/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(restConfig))
	http.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("swagger-ui/dist"))))
	http.Handle("/swagger/", http.RedirectHandler("/swagger-ui/?url=/apidocs.json", http.StatusFound))
	logrus.Infof("ready to serve on port %d", setting.Config.HttpConfig.HTTPPort)

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

	// make sure worker starts after all tasks registered
	task.StartWorker()

	//shut down gracefully
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	err = server.Shutdown(context.Background())
	if err != nil {
		logrus.Error(err.Error())
	}
	close(stopChan)
	task.StopWorker(30)
	logrus.Info("waiting for informer stopping")
	time.Sleep(2 * time.Second)
	logrus.Info("walm server stopped gracefully")
	return nil
}

func RouteLogging(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	now := time.Now()
	chain.ProcessFilter(req, resp)
	logrus.Infof("[route-filter (logger)] CLIENT %s OP %s URI %s COST %v RESP %d", req.Request.Host, req.Request.Method, req.Request.URL, time.Now().Sub(now), resp.StatusCode())
}

var ServerStats = stats.New()

func ServerStatsFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	beginning, recorder := ServerStats.Begin(response)
	chain.ProcessFilter(request, response)
	ServerStats.End(beginning, stats.WithRecorder(recorder))
}

func ServerStatsData(request *restful.Request, response *restful.Response) {
	response.WriteEntity(ServerStats.Data())
}

func readinessProbe(request *restful.Request, response *restful.Response) {
	response.WriteEntity("OK")
}

func livenessProbe(request *restful.Request, response *restful.Response) {
	response.WriteEntity("OK")
}

func InitRootRouter() *restful.WebService {
	ws := new(restful.WebService)

	ws.Path("/").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	tags := []string{"root"}

	ws.Route(ws.GET("/readiness").To(readinessProbe).
		Doc("服务Ready状态检查").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", httpModel.ErrorMessageResponse{}))

	ws.Route(ws.GET("/liveniess").To(livenessProbe).
		Doc("服务Live状态检查").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", httpModel.ErrorMessageResponse{}))

	ws.Route(ws.GET("/stats").To(ServerStatsData).
		Doc("获取服务Stats").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Returns(200, "OK", nil).
		Returns(500, "Internal Error", httpModel.ErrorMessageResponse{}))

	return ws
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

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Walm",
			Description: "Walm Web Server",
			Version:     "0.0.1",
		},
	}
}
