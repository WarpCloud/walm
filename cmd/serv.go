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
	"walm/pkg/setting"
	"walm/pkg/k8s/informer"
	"walm/pkg/release/manager/helm"
)

const servDesc = `
This command enable a WALM Web server.

$ walm serv 

Before to start serv ,you need to config the conf file 

The file is named conf.yaml

`

type ServCmd struct {
	cfgFile string
}

func initService() error {
	informer.InitInformer()
	helm.InitHelm()

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

	// accept and respond in JSON unless told otherwise
	restful.DefaultRequestContentType(restful.MIME_JSON)
	restful.DefaultResponseContentType(restful.MIME_JSON)
	// gzip if accepted
	restful.DefaultContainer.EnableContentEncoding(true)
	// faster router
	restful.DefaultContainer.Router(restful.CurlyRouter{})

	logrus.Infoln("Adding Route...")
	restful.Add(router.InitRootRouter())
	restful.Add(router.InitNodeRouter())
	restful.Add(router.InitTenantRouter())
	restful.Add(router.InitInstanceRouter())
	restful.Add(router.InitClusterRouter())
	restful.Add(router.InitPodRouter())
	logrus.Infoln("Add Route Success")

	config := restfulspec.Config{
		// You control what services are visible
		WebServices:    restful.RegisteredWebServices(),
		APIPath:        "/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	http.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir("swagger-ui/dist"))))

	logrus.Infoln("ready to serve on")
	server := &http.Server{Addr: fmt.Sprintf(":%d", setting.Config.HttpConfig.HTTPPort), Handler: restful.DefaultContainer}
	logrus.Fatalln(server.ListenAndServe())
	// server.ListenAndServeTLS()

	return nil
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Walm",
			Description: "Walm Web Server",
			Version: "0.0.1",
		},
	}
}
