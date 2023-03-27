package helm

import (
	"WarpCloud/walm/pkg/helm/impl"
	"WarpCloud/walm/pkg/k8s/cache/informer"
	"WarpCloud/walm/pkg/models/common"
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/setting"
	"WarpCloud/walm/pkg/util"
	"WarpCloud/walm/test/e2e/framework"
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"path/filepath"
	"strings"

	"WarpCloud/walm/pkg/helm/impl/plugins"
	"WarpCloud/walm/pkg/models/k8s"
)

var _ = Describe("HelmRelease", func() {

	var (
		namespace string
		helm      *impl.Helm
		err       error
		stopChan  chan struct{}
	)

	BeforeEach(func() {
		By("create namespace")
		namespace, err = framework.CreateRandomNamespace("helmReleaseTest", nil)
		Expect(err).NotTo(HaveOccurred())
		stopChan = make(chan struct{})
		k8sCache := informer.NewInformer(framework.GetK8sClient(), framework.GetK8sReleaseConfigClient(), framework.GetK8sInstanceClient(), nil, nil, nil, 0, stopChan)
		registryClient, err := impl.NewRegistryClient(setting.Config.ChartImageConfig)
		Expect(err).NotTo(HaveOccurred())

		helm, err = impl.NewHelm(setting.Config.RepoList, registryClient, k8sCache, framework.GetKubeClient())
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		By("delete namespace")
		close(stopChan)
		err = framework.DeleteNamespace(namespace)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("test install release", func() {

		var (
			tomcatChartFiles     []*common.BufferedFile
			tomcatComputedValues map[string]interface{}
		)

		BeforeEach(func() {
			tomcatChartPath, err := framework.GetLocalTomcatChartPath()
			Expect(err).NotTo(HaveOccurred())

			tomcatChartFiles, err = framework.LoadChartArchive(tomcatChartPath)
			Expect(err).NotTo(HaveOccurred())

			defaultValues, err := getChartDefaultValues(tomcatChartFiles)
			Expect(err).NotTo(HaveOccurred())
			tomcatComputedValues = map[string]interface{}{}
			tomcatComputedValues = util.MergeValues(tomcatComputedValues, defaultValues, false)
			tomcatComputedValues = util.MergeValues(tomcatComputedValues, map[string]interface{}{
				plugins.WalmPluginConfigKey: []*k8s.ReleasePlugin{
					{
						Name: plugins.ValidateReleaseConfigPluginName,
					},
					{
						Name: plugins.IsomateSetConverterPluginName,
					},
				},
			}, false)
		})

		It("test release local chart", func() {
			By("install release with local chart")
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "tomcat-test",
				},
			}

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, tomcatChartFiles, false, false, nil)
			Expect(err).NotTo(HaveOccurred())
			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 1)

			Expect(releaseCache.ComputedValues).To(Equal(tomcatComputedValues))

			manifest := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartImage: \"\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  chartWalmVersion: v2\n  configValues: {}\n  dependencies: null\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n",
				namespace, namespace, namespace)
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

			mataInfoParams := getTomcatChartDefaultMetaInfoParams()
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})

		It("test release repo chart", func() {
			By("install release with repo chart")
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:         "tomcat-test",
					RepoName:     framework.TestChartRepoName,
					ChartName:    framework.TomcatChartName,
					ChartVersion: framework.TomcatChartVersion,
				},
			}

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, nil, false, false, nil)
			Expect(err).NotTo(HaveOccurred())
			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 1)

			Expect(releaseCache.ComputedValues).To(Equal(tomcatComputedValues))

			manifest := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartImage: \"\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  chartWalmVersion: v2\n  configValues: {}\n  dependencies: null\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig: {}\n  repo: %s\nstatus: {}\n",
				namespace, namespace, namespace, framework.TestChartRepoName)
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

			mataInfoParams := getTomcatChartDefaultMetaInfoParams()
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})

		It("test release chart image", func() {
			By("install release with chart image")
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "tomcat-test",
				},
				ChartImage: framework.GetTomcatChartImage(),
			}

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, nil, false, false, nil)
			Expect(err).NotTo(HaveOccurred())
			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 1)

			Expect(releaseCache.ComputedValues).To(Equal(tomcatComputedValues))

			manifest := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartImage: %s\n  chartName: tomcat\n  chartVersion: 0.2.0\n  chartWalmVersion: v2\n  configValues: {}\n  dependencies: null\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n",
				namespace, namespace, namespace, framework.GetTomcatChartImage())
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

			mataInfoParams := getTomcatChartDefaultMetaInfoParams()
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})

		It("test release jsonnet chart", func() {
			By("install release with jsonnet chart")
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "zookeeper-test",
				},
			}

			currentFilePath, err := framework.GetCurrentFilePath()
			Expect(err).NotTo(HaveOccurred())

			chartPath := filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/zookeeper-6.1.0.tgz")
			Expect(err).NotTo(HaveOccurred())

			chartFiles, err := framework.LoadChartArchive(chartPath)
			Expect(err).NotTo(HaveOccurred())

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, chartFiles, false, false, nil)
			Expect(err).NotTo(HaveOccurred())
			assertReleaseCacheBasic(releaseCache, namespace, "zookeeper-test", "", "zookeeper",
				"6.1.0", "6.1", 1)

			defaultValues, err := getChartDefaultValues(chartFiles)
			Expect(err).NotTo(HaveOccurred())
			computedValues := map[string]interface{}{}
			computedValues = util.MergeValues(computedValues, defaultValues, false)
			computedValues = util.MergeValues(computedValues, map[string]interface{}{
				plugins.WalmPluginConfigKey: []*k8s.ReleasePlugin{
					{
						Name: plugins.ValidateReleaseConfigPluginName,
					},
					{
						Name: plugins.IsomateSetConverterPluginName,
					},
				},
			}, false)

			Expect(releaseCache.ComputedValues).To(Equal(computedValues))

			manifest := strings.Replace("\n---\napiVersion: v1\ndata:\n  jaas.conf.tmpl: |\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    Server {\n      com.sun.security.auth.module.Krb5LoginModule required\n      useKeyTab=true\n      keyTab=\"/etc/keytabs/keytab\"\n      storeKey=true\n      useTicketCache=false\n      principal=\"{{ getv \"/security/guardian_principal_user\" \"zookeeper\" }}/{{ getv \"/security/guardian_principal_host\" \"tos\" }}@{{ getv \"/security/guardian_client_config/realm\" \"TDH\" }}\";\n    };\n    Client {\n      com.sun.security.auth.module.Krb5LoginModule required\n      useKeyTab=false\n      useTicketCache=true;\n    };\n    {{- end }}\n  log4j.properties.raw: |\n    # Define some default values that can be overridden by system properties\n    zookeeper.root.logger=INFO, CONSOLE\n    zookeeper.console.threshold=INFO\n    zookeeper.log.dir=.\n    zookeeper.log.file=zookeeper.log\n    zookeeper.log.threshold=DEBUG\n    zookeeper.tracelog.dir=.\n    zookeeper.tracelog.file=zookeeper_trace.log\n\n    #\n    # ZooKeeper Logging Configuration\n    #\n\n    # Format is \"<default threshold> (, <appender>)+\n\n    # DEFAULT: console appender only\n    log4j.rootLogger=${zookeeper.root.logger}\n\n    # Example with rolling log file\n    #log4j.rootLogger=DEBUG, CONSOLE, ROLLINGFILE\n\n    # Example with rolling log file and tracing\n    #log4j.rootLogger=TRACE, CONSOLE, ROLLINGFILE, TRACEFILE\n\n    #\n    # Log INFO level and above messages to the console\n    #\n    log4j.appender.CONSOLE=org.apache.log4j.ConsoleAppender\n    log4j.appender.CONSOLE.Threshold=${zookeeper.log.threshold}\n    log4j.appender.CONSOLE.layout=org.apache.log4j.PatternLayout\n    log4j.appender.CONSOLE.layout.ConversionPattern=%d{ISO8601} %-5p %c: [myid:%X{myid}] - [%t:%C{1}@%L] - %m%n\n\n    #\n    # Add ROLLINGFILE to rootLogger to get log file output\n    #    Log DEBUG level and above messages to a log file\n    log4j.appender.ROLLINGFILE=org.apache.log4j.RollingFileAppender\n    log4j.appender.ROLLINGFILE.Threshold=${zookeeper.log.threshold}\n    log4j.appender.ROLLINGFILE.File=${zookeeper.log.dir}/${zookeeper.log.file}\n\n    # Max log file size of 10MB\n    log4j.appender.ROLLINGFILE.MaxFileSize=64MB\n    # uncomment the next line to limit number of backup files\n    log4j.appender.ROLLINGFILE.MaxBackupIndex=4\n\n    log4j.appender.ROLLINGFILE.layout=org.apache.log4j.PatternLayout\n    log4j.appender.ROLLINGFILE.layout.ConversionPattern=%d{ISO8601} %-5p %c: [myid:%X{myid}] - [%t:%C{1}@%L] - %m%n\n\n\n    #\n    # Add TRACEFILE to rootLogger to get log file output\n    #    Log DEBUG level and above messages to a log file\n    log4j.appender.TRACEFILE=org.apache.log4j.FileAppender\n    log4j.appender.TRACEFILE.Threshold=TRACE\n    log4j.appender.TRACEFILE.File=${zookeeper.tracelog.dir}/${zookeeper.tracelog.file}\n\n    log4j.appender.TRACEFILE.layout=org.apache.log4j.PatternLayout\n    ### Notice we are including log4j's NDC here (%x)\n    log4j.appender.TRACEFILE.layout.ConversionPattern=%d{ISO8601} %-5p %c: [myid:%X{myid}] - [%t:%C{1}@%L][%x] - %m%n\n  myid.tmpl: '{{ getenv \"MYID\" }}'\n  tdh-env.sh.tmpl: |\n    #!/bin/bash\n    set -x\n\n    setup_keytab() {\n      echo \"setup_keytab\"\n    {{ if eq (getv \"/security/auth_type\") \"kerberos\" }}\n      # link_keytab\n      export KRB_MOUNTED_CONF_PATH=${KRB_MOUNTED_CONF_PATH:-/var/run/secrets/transwarp.io/tosvolume/keytab/krb5.conf}\n      export KRB_MOUNTED_KEYTAB=${KRB_MOUNTED_KEYTAB:-/var/run/secrets/transwarp.io/tosvolume/keytab/keytab}\n      if [ ! -f $KRB_MOUNTED_CONF_PATH ]; then\n        echo \"Expect krb5.conf at $KRB_MOUNTED_CONF_PATH but not found!\"\n        exit 1\n      fi\n      if [ ! -f $KRB_MOUNTED_KEYTAB ]; then\n        echo \"Expect keytab file at $KRB_MOUNTED_KEYTAB but not found!\"\n        exit 1\n      fi\n      ln -svf $KRB_MOUNTED_CONF_PATH /etc/krb5.conf\n      [ -d /etc/keytabs ] || mkdir -p /etc/keytabs\n      ln -svf $KRB_MOUNTED_KEYTAB /etc/keytabs/keytab\n    {{ end }}\n    }\n  tdh-env.toml: |-\n    [[template]]\n    src = \"tdh-env.sh.tmpl\"\n    dest = \"/etc/tdh-env.sh\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n  zoo.cfg.tmpl: |\n    # the directory where the snapshot is stored.\n    dataDir=/var/transwarp/data\n\n    # the port at which the clients will connect\n    clientPort={{ getv \"/zookeeper/zookeeper.client.port\" }}\n\n    {{- range $index, $_ := seq 0 (sub (atoi (getenv \"QUORUM_SIZE\")) 1) }}\n    server.{{ $index }}={{ getenv \"SERVICE_NAME\" }}-{{ $index }}.{{ getenv \"SERVICE_NAME\" }}-hl.{{ getenv \"SERVICE_NAMESPACE\" }}.svc:{{ getv \"/zookeeper/zookeeper.peer.communicate.port\" }}:{{ getv \"/zookeeper/zookeeper.leader.elect.port\" }}\n    {{- end }}\n\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    authProvider.1=org.apache.zookeeper.server.auth.SASLAuthenticationProvider\n    jaasLoginRenew=3600000\n    kerberos.removeHostFromPrincipal=true\n    kerberos.removeRealmFromPrincipal=true\n    {{- end }}\n\n    {{- range gets \"/zoo_cfg/*\" }}\n    {{base .Key}}={{.Value}}\n    {{- end }}\n  zookeeper-confd.conf: |-\n    {\n      \"security\": {\n        \"auth_type\": \"none\"\n      },\n      \"transwarpApplicationPause\": false,\n      \"transwarpCniNetwork\": \"overlay\",\n      \"transwarpGlobalIngress\": {\n        \"httpPort\": 80,\n        \"httpsPort\": 443\n      },\n      \"transwarpLicenseAddress\": \"\",\n      \"transwarpMetrics\": {\n        \"enable\": true\n      },\n      \"zoo_cfg\": {\n        \"autopurge.purgeInterval\": 5,\n        \"autopurge.snapRetainCount\": 10,\n        \"initLimit\": 10,\n        \"maxClientCnxns\": 0,\n        \"syncLimit\": 5,\n        \"tickTime\": 9000\n      },\n      \"zookeeper\": {\n        \"zookeeper.client.port\": 2181,\n        \"zookeeper.jmxremote.port\": 9911,\n        \"zookeeper.leader.elect.port\": 3888,\n        \"zookeeper.peer.communicate.port\": 2888\n      }\n    }\n  zookeeper-env.sh.tmpl: |\n    export ZOOKEEPER_LOG_DIR=/var/transwarp/data/log\n\n    export SERVER_JVMFLAGS=\"-Dcom.sun.management.jmxremote.port={{getv \"/zookeeper/zookeeper.jmxremote.port\"}} -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.local.only=false\"\n    export SERVER_JVMFLAGS=\"-Dsun.net.inetaddr.ttl=60 -Dsun.net.inetaddr.negative.ttl=60 -Dzookeeper.refreshPeer=1 -Dzookeeper.log.dir=${ZOOKEEPER_LOG_DIR} -Dzookeeper.root.logger=INFO,CONSOLE,ROLLINGFILE $SERVER_JVMFLAGS\"\n\n    {{ if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    export SERVER_JVMFLAGS=\"-Djava.security.auth.login.config=/etc/zookeeper/conf/jaas.conf ${SERVER_JVMFLAGS}\"\n    export ZOOKEEPER_PRICIPAL={{ getv \"/security/guardian_principal_user\" \"zookeeper\" }}/{{ getv \"/security/guardian_principal_host\" \"tos\" }}@{{ getv \"/security/guardian_client_config/realm\" \"TDH\" }}\n    {{ end }}\n  zookeeper.toml: |-\n    [[template]]\n    src = \"zoo.cfg.tmpl\"\n    dest = \"/etc/zookeeper/conf/zoo.cfg\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"jaas.conf.tmpl\"\n    dest = \"/etc/zookeeper/conf/jaas.conf\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"log4j.properties.raw\"\n    dest = \"/etc/zookeeper/conf/log4j.properties\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"zookeeper-env.sh.tmpl\"\n    dest = \"/etc/zookeeper/conf/zookeeper-env.sh\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"myid.tmpl\"\n    dest = \"/var/transwarp/data/myid\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\nkind: ConfigMap\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: zookeeper\n    app.kubernetes.io/instance: zookeeper-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: zookeeper-confd-conf\n    app.kubernetes.io/part-of: zookeeper\n    app.kubernetes.io/version: \"6.1\"\n  name: zookeeper-test-zookeeper-confd-conf\n  namespace: helmreleasetest-67t9g\n\n---\napiVersion: v1\ndata:\n  entrypoint.sh: |\n    #!/bin/bash\n    set -ex\n\n    export ZOOKEEPER_CONF_DIR=/etc/zookeeper/conf\n    export ZOOKEEPER_DATA_DIR=/var/transwarp\n    export ZOOKEEPER_DATA=$ZOOKEEPER_DATA_DIR/data\n    export ZOOKEEPER_CFG=$ZOOKEEPER_CONF_DIR/zoo.cfg\n\n    mkdir -p ${ZOOKEEPER_CONF_DIR}\n    mkdir -p $ZOOKEEPER_DATA\n\n    export MYID=${HOSTNAME##*-}\n    confd -onetime -backend file -prefix / -file /etc/confd/zookeeper-confd.conf\n\n    ZOOKEEPER_ENV=$ZOOKEEPER_CONF_DIR/zookeeper-env.sh\n\n    [ -f $ZOOKEEPER_ENV ] && {\n      source $ZOOKEEPER_ENV\n    }\n    [ -f /etc/tdh-env.sh ] && {\n      source /etc/tdh-env.sh\n      setup_keytab\n    }\n    # ZOOKEEPER_LOG is defined in $ZOOKEEPER_ENV\n    mkdir -p $ZOOKEEPER_LOG_DIR\n    chown -R zookeeper:zookeeper $ZOOKEEPER_LOG_DIR\n    chown -R zookeeper:zookeeper $ZOOKEEPER_DATA\n\n    echo \"Starting zookeeper service with config:\"\n    cat ${ZOOKEEPER_CFG}\n\n    JMXEXPORTER_ENABLED=${JMXEXPORTER_ENABLED:-\"true\"}\n    if [ \"${JMXEXPORTER_ENABLED}\" == \"true\" ];then\n      export JAVAAGENT_OPTS=\" -javaagent:/usr/lib/jmx_exporter/jmx_prometheus_javaagent-0.7.jar=19000:/usr/lib/jmx_exporter/agentconfig.yml \"\n    fi\n\n    sudo -u zookeeper java $SERVER_JVMFLAGS \\\n        $JAVAAGENT_OPTS \\\n        -cp $ZOOKEEPER_HOME/zookeeper-3.4.5-transwarp-with-dependencies.jar:$ZOOKEEPER_CONF_DIR \\\n        org.apache.zookeeper.server.quorum.QuorumPeerMain $ZOOKEEPER_CFG\nkind: ConfigMap\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: zookeeper\n    app.kubernetes.io/instance: zookeeper-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: zookeeper-entrypoint\n    app.kubernetes.io/part-of: zookeeper\n    app.kubernetes.io/version: \"6.1\"\n  name: zookeeper-test-zookeeper-entrypoint\n  namespace: helmreleasetest-67t9g\n\n---\napiVersion: v1\nkind: Service\nmetadata:\n  annotations:\n    service.alpha.kubernetes.io/tolerate-unready-endpoints: \"true\"\n  labels:\n    app.kubernetes.io/component: zookeeper\n    app.kubernetes.io/instance: zookeeper-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: zookeeper\n    app.kubernetes.io/part-of: zookeeper\n    app.kubernetes.io/service-type: headless-service\n    app.kubernetes.io/version: \"6.1\"\n  name: zookeeper-test-zookeeper-hl\n  namespace: helmreleasetest-67t9g\nspec:\n  clusterIP: None\n  ports:\n  - name: zk-port\n    port: 2181\n    targetPort: 2181\n  selector:\n    app.kubernetes.io/instance: zookeeper-test\n    app.kubernetes.io/name: zookeeper\n    app.kubernetes.io/version: \"6.1\"\n\n---\napiVersion: apps/v1beta1\nkind: StatefulSet\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: zookeeper\n    app.kubernetes.io/instance: zookeeper-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: zookeeper\n    app.kubernetes.io/part-of: zookeeper\n    app.kubernetes.io/version: \"6.1\"\n  name: zookeeper-test-zookeeper\n  namespace: helmreleasetest-67t9g\nspec:\n  podManagementPolicy: Parallel\n  replicas: 3\n  selector:\n    matchLabels:\n      app.kubernetes.io/instance: zookeeper-test\n      app.kubernetes.io/name: zookeeper\n      app.kubernetes.io/version: \"6.1\"\n  serviceName: zookeeper-test-zookeeper-hl\n  template:\n    metadata:\n      annotations:\n        tos.network.staticIP: \"true\"\n      labels:\n        app.kubernetes.io/instance: zookeeper-test\n        app.kubernetes.io/name: zookeeper\n        app.kubernetes.io/version: \"6.1\"\n    spec:\n      affinity:\n        podAntiAffinity:\n          requiredDuringSchedulingIgnoredDuringExecution:\n          - labelSelector:\n              matchLabels:\n                app.kubernetes.io/instance: zookeeper-test\n                app.kubernetes.io/name: zookeeper\n                app.kubernetes.io/version: \"6.1\"\n            namespaces:\n            - helmreleasetest-67t9g\n            topologyKey: kubernetes.io/hostname\n      containers:\n      - command:\n        - /boot/entrypoint.sh\n        env:\n        - name: SERVICE_NAME\n          value: zookeeper-test-zookeeper\n        - name: SERVICE_NAMESPACE\n          value: helmreleasetest-67t9g\n        - name: QUORUM_SIZE\n          value: \"3\"\n        image: docker.io/corndai1997/zookeeper:5.2\n        imagePullPolicy: Always\n        name: zookeeper\n        readinessProbe:\n          exec:\n            command:\n            - /bin/bash\n            - -c\n            - echo ruok|nc localhost 2181 > /dev/null && echo ok\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: \"0.20000000000000001\"\n            memory: 200Mi\n            nvidia.com/gpu: \"0\"\n          requests:\n            cpu: \"0.10000000000000001\"\n            memory: 100Mi\n            nvidia.com/gpu: \"0\"\n        volumeMounts:\n        - mountPath: /boot\n          name: zookeeper-entrypoint\n        - mountPath: /etc/confd\n          name: zookeeper-confd-conf\n        - mountPath: /var/transwarp\n          name: zkdir\n      hostNetwork: false\n      initContainers: []\n      priorityClassName: low-priority\n      restartPolicy: Always\n      terminationGracePeriodSeconds: 30\n      volumes:\n      - configMap:\n          items:\n          - key: entrypoint.sh\n            mode: 493\n            path: entrypoint.sh\n          name: zookeeper-test-zookeeper-entrypoint\n        name: zookeeper-entrypoint\n      - configMap:\n          items:\n          - key: zookeeper.toml\n            path: conf.d/zookeeper.toml\n          - key: tdh-env.toml\n            path: conf.d/tdh-env.toml\n          - key: zookeeper-confd.conf\n            path: zookeeper-confd.conf\n          - key: zoo.cfg.tmpl\n            path: templates/zoo.cfg.tmpl\n          - key: jaas.conf.tmpl\n            path: templates/jaas.conf.tmpl\n          - key: zookeeper-env.sh.tmpl\n            path: templates/zookeeper-env.sh.tmpl\n          - key: myid.tmpl\n            path: templates/myid.tmpl\n          - key: log4j.properties.raw\n            path: templates/log4j.properties.raw\n          - key: tdh-env.sh.tmpl\n            path: templates/tdh-env.sh.tmpl\n          name: zookeeper-test-zookeeper-confd-conf\n        name: zookeeper-confd-conf\n  updateStrategy:\n    type: RollingUpdate\n  volumeClaimTemplates:\n  - metadata:\n      annotations:\n        volume.beta.kubernetes.io/storage-class: local\n      labels:\n        app.kubernetes.io/component: zookeeper\n        app.kubernetes.io/instance: zookeeper-test\n        app.kubernetes.io/managed-by: walm\n        app.kubernetes.io/name: zookeeper\n        app.kubernetes.io/part-of: zookeeper\n        app.kubernetes.io/version: \"6.1\"\n      name: zkdir\n    spec:\n      accessModes:\n      - ReadWriteOnce\n      resources:\n        requests:\n          storage: 100Gi\n      storageClassName: local\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  labels:\n    app.kubernetes.io/component: zookeeper\n    app.kubernetes.io/instance: zookeeper-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: zookeeper\n    app.kubernetes.io/part-of: zookeeper\n    app.kubernetes.io/version: \"6.1\"\n  name: zookeeper-test\n  namespace: helmreleasetest-67t9g\nspec:\n  chartAppVersion: \"6.1\"\n  chartImage: \"\"\n  chartName: zookeeper\n  chartVersion: 6.1.0\n  chartWalmVersion: v2\n  configValues: {}\n  dependencies: null\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig:\n    zookeeper_addresses: zookeeper-test-zookeeper-0.zookeeper-test-zookeeper-hl.helmreleasetest-67t9g.svc,zookeeper-test-zookeeper-1.zookeeper-test-zookeeper-hl.helmreleasetest-67t9g.svc,zookeeper-test-zookeeper-2.zookeeper-test-zookeeper-hl.helmreleasetest-67t9g.svc\n    zookeeper_auth_type: none\n    zookeeper_port: \"2181\"\n  repo: \"\"\nstatus: {}\n",
				"helmreleasetest-67t9g", namespace, -1)
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getZookeeperDefaultReleaseResourceMeta(namespace, "zookeeper-test")))

			mataInfoParams := getZookeeperDefaultMetaInfoParams()
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})

		It("test release dry run", func() {
			By("install release by dry run")
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "tomcat-test",
				},
			}

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, tomcatChartFiles, true, false, nil)
			Expect(err).NotTo(HaveOccurred())
			manifest := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartImage: \"\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  chartWalmVersion: v2\n  configValues: {}\n  dependencies: null\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n",
				namespace, namespace, namespace)
			Expect(releaseCache.Manifest).To(Equal(manifest))
		})

		It("test release metainfo params", func() {
			By("install release with metainfo params")
			replicas := int64(2)
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "tomcat-test",
				},
				MetaInfoParams: &release.MetaInfoParams{
					Roles: []*release.MetaRoleConfigValue{
						{
							Name: "tomcat",
							RoleBaseConfigValue: &release.MetaRoleBaseConfigValue{
								Replicas: &replicas,
							},
						},
					},
				},
			}

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, tomcatChartFiles, false, false, nil)
			Expect(err).NotTo(HaveOccurred())
			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 1)

			tomcatComputedValues["replicaCount"] = 2
			assertYamlConfigValues(releaseCache.ComputedValues, tomcatComputedValues)

			manifest := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 2\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartImage: \"\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  chartWalmVersion: v2\n  configValues:\n    replicaCount: 2\n  dependencies: null\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n",
				namespace, namespace, namespace)
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

			mataInfoParams := getTomcatChartDefaultMetaInfoParams()
			mataInfoParams.Roles[1].RoleBaseConfigValue.Replicas = &replicas
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})

		It("test release label", func() {
			By("install release with label")
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "tomcat-test",
				},
				ReleaseLabels: map[string]string{"walm-test": "true"},
			}

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, tomcatChartFiles, false, false, nil)
			Expect(err).NotTo(HaveOccurred())
			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 1)

			assertYamlConfigValues(releaseCache.ComputedValues, tomcatComputedValues)

			manifest := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  labels:\n    walm-test: \"true\"\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartImage: \"\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  chartWalmVersion: v2\n  configValues: {}\n  dependencies: null\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n",
				namespace, namespace, namespace)
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

			mataInfoParams := getTomcatChartDefaultMetaInfoParams()
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})

		It("test release plugin", func() {
			By("install release with plugin")
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "tomcat-test",
				},
				Plugins: []*k8s.ReleasePlugin{
					{
						Name: plugins.LabelPodPluginName,
						Args: "{\"labelsToAdd\":{\"test_key\":\"test_value\"}}",
					},
				},
			}

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, tomcatChartFiles, false, false, nil)
			Expect(err).NotTo(HaveOccurred())
			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 1)

			tomcatComputedValues = util.MergeValues(tomcatComputedValues, map[string]interface{}{
				plugins.WalmPluginConfigKey: []*k8s.ReleasePlugin{
					{
						Name: plugins.LabelPodPluginName,
						Args: "{\"labelsToAdd\":{\"test_key\":\"test_value\"}}",
					},
					{
						Name: plugins.ValidateReleaseConfigPluginName,
					},
					{
						Name: plugins.IsomateSetConverterPluginName,
					},
				},
			}, false)
			assertYamlConfigValues(releaseCache.ComputedValues, tomcatComputedValues)

			manifest := strings.Replace("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: helmreleasetest-gck8q\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: helmreleasetest-gck8q\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n        test_key: test_value\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: helmreleasetest-gck8q\nspec:\n  chartAppVersion: \"7\"\n  chartImage: \"\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  chartWalmVersion: v2\n  configValues: {}\n  dependencies: null\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n",
				"helmreleasetest-gck8q", namespace, -1)
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

			mataInfoParams := getTomcatChartDefaultMetaInfoParams()
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})

		// kafka manifest contains annotation md5
		Describe("need fixed namespaces", func() {
			var (
				fixedNamespace1 string
				fixedNamespace2 string
			)

			BeforeEach(func() {
				By("create fixed namespace")
				fixedNamespace1 = "helmreleasetest-fixedns1"
				err = framework.CreateNamespace(fixedNamespace1, nil)
				Expect(err).NotTo(HaveOccurred())
				fixedNamespace2 = "helmreleasetest-fixedns2"
				err = framework.CreateNamespace(fixedNamespace2, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				By("delete fixed namespace")
				err = framework.DeleteNamespace(fixedNamespace1)
				Expect(err).NotTo(HaveOccurred())
				err = framework.DeleteNamespace(fixedNamespace2)
				Expect(err).NotTo(HaveOccurred())
			})

			It("test release dependency", func() {
				By("install zookeeper")
				releaseRequest := &release.ReleaseRequestV2{
					ReleaseRequest: release.ReleaseRequest{
						Name: "zookeeper-test",
					},
				}

				currentFilePath, err := framework.GetCurrentFilePath()
				Expect(err).NotTo(HaveOccurred())

				chartPath := filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/zookeeper-6.1.0.tgz")
				Expect(err).NotTo(HaveOccurred())

				chartFiles, err := framework.LoadChartArchive(chartPath)
				Expect(err).NotTo(HaveOccurred())

				_, err = helm.InstallOrCreateRelease(fixedNamespace1, releaseRequest, chartFiles, false, false, nil)
				Expect(err).NotTo(HaveOccurred())

				By("install kafka which depends on zookeeper")
				releaseRequest = &release.ReleaseRequestV2{
					ReleaseRequest: release.ReleaseRequest{
						Name:         "kafka-test",
						Dependencies: map[string]string{"zookeeper": "zookeeper-test-notexist"},
					},
				}

				currentFilePath, err = framework.GetCurrentFilePath()
				Expect(err).NotTo(HaveOccurred())

				chartPath = filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/kafka-6.1.0.tgz")
				Expect(err).NotTo(HaveOccurred())

				chartFiles, err = framework.LoadChartArchive(chartPath)
				Expect(err).NotTo(HaveOccurred())

				_, err = helm.InstallOrCreateRelease(fixedNamespace1, releaseRequest, chartFiles, false, false, nil)
				Expect(err).To(HaveOccurred())

				releaseRequest.Dependencies = map[string]string{"zookeeper": "zookeeper-test"}
				releaseCache, err := helm.InstallOrCreateRelease(fixedNamespace1, releaseRequest, chartFiles, false, false, nil)
				Expect(err).NotTo(HaveOccurred())

				assertReleaseCacheBasic(releaseCache, fixedNamespace1, "kafka-test", "", "kafka",
					"6.1.0", "6.1", 1)

				defaultValues, err := getChartDefaultValues(chartFiles)
				Expect(err).NotTo(HaveOccurred())
				computedValues := map[string]interface{}{}
				computedValues = util.MergeValues(computedValues, defaultValues, false)
				computedValues = util.MergeValues(computedValues, map[string]interface{}{
					plugins.WalmPluginConfigKey: []*k8s.ReleasePlugin{
						{
							Name: plugins.ValidateReleaseConfigPluginName,
						},
						{
							Name: plugins.IsomateSetConverterPluginName,
						},
					},
					"ZOOKEEPER_CLIENT_CONFIG": map[string]interface{}{
						"zookeeper_addresses": "zookeeper-test-zookeeper-0.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc,zookeeper-test-zookeeper-1.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc,zookeeper-test-zookeeper-2.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc",
						"zookeeper_auth_type": "none",
						"zookeeper_port":      "2181",
					},
				}, false)

				Expect(releaseCache.ComputedValues).To(Equal(computedValues))

				manifest := "\n---\napiVersion: v1\ndata:\n  consumer.properties.tmpl: |-\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" -}}\n    bootstrap.servers={{getenv \"KAFKA_HOSTNAME\" \"localhost\"}}:9092\n    sasl.mechanism=GSSAPI\n    security.protocol=SASL_PLAINTEXT\n    sasl.kerberos.service.name={{ getv \"/security/guardian_principal_user\" \"kafka\" }}\n    sasl.kerberos.service.principal.instance={{ getv \"/security/guardian_principal_host\" \"tos\" }}\n    {{- else }}\n    bootstrap.servers={{getenv \"KAFKA_HOSTNAME\" \"localhost\"}}:9092\n    security.protocol=PLAINTEXT\n    sasl.mechanism=PLAIN\n    {{- end }}\n  jaas.conf.tmpl: |\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    KafkaServer {\n      com.sun.security.auth.module.Krb5LoginModule required\n      useKeyTab=true\n      keyTab=\"/etc/keytabs/keytab\"\n      storeKey=true\n      useTicketCache=false\n      principal=\"{{ getv \"/security/guardian_principal_user\" \"kafka\" }}/{{ getv \"/security/guardian_principal_host\" \"tos\" }}@{{ getv \"/security/guardian_client_config/realm\" \"TDH\" }}\";\n    };\n    KafkaClient {\n      com.sun.security.auth.module.Krb5LoginModule required\n      useKeyTab=true\n      keyTab=\"/etc/keytabs/keytab\"\n      storeKey=true\n      useTicketCache=false\n      principal=\"{{ getv \"/security/guardian_principal_user\" \"kafka\" }}/{{ getv \"/security/guardian_principal_host\" \"tos\" }}@{{ getv \"/security/guardian_client_config/realm\" \"TDH\" }}\";\n    };\n    // Zookeeper client authentication\n    Client {\n      com.sun.security.auth.module.Krb5LoginModule required\n      useKeyTab=true\n      storeKey=true\n      useTicketCache=false\n      keyTab=\"/etc/keytabs/keytab\"\n      principal=\"{{ getv \"/security/guardian_principal_user\" \"kafka\" }}/{{ getv \"/security/guardian_principal_host\" \"tos\" }}@{{ getv \"/security/guardian_client_config/realm\" \"TDH\" }}\";\n    };\n    {{- end }}\n  kafka-confd.conf: |-\n    {\n      \"java_opts\": {\n        \"memory_opts\": {\n          \"kafka_memory\": \"3276\"\n        }\n      },\n      \"kafka\": {\n\n      },\n      \"security\": {\n        \"auth_type\": \"none\"\n      },\n      \"server_properties\": {\n        \"default.replication.factor\": 2,\n        \"log.dirs\": \"/data\",\n        \"log.flush.interval.messages\": 10000,\n        \"log.flush.interval.ms\": 1000,\n        \"log.retention.bytes\": 1073741824,\n        \"log.retention.check.interval.ms\": 300000,\n        \"log.retention.hours\": 6,\n        \"log.segment.bytes\": 1073741824,\n        \"message.max.bytes\": 100000000,\n        \"num.io.threads\": 8,\n        \"num.network.threads\": 3,\n        \"num.partitions\": 3,\n        \"num.recovery.threads.per.data.dir\": 1,\n        \"replica.fetch.max.bytes\": 100000000,\n        \"socket.receive.buffer.bytes\": 102400,\n        \"socket.request.max.bytes\": 104857600,\n        \"socket.send.buffer.bytes\": 102400,\n        \"zookeeper.connection.timeout.ms\": 6000\n      },\n      \"transwarpApplicationPause\": false,\n      \"transwarpCniNetwork\": \"overlay\",\n      \"transwarpGlobalIngress\": {\n        \"httpPort\": 80,\n        \"httpsPort\": 443\n      },\n      \"transwarpLicenseAddress\": \"\",\n      \"transwarpMetrics\": {\n        \"enable\": true\n      },\n      \"zookeeper_client_config\": {\n        \"zookeeper_addresses\": \"zookeeper-test-zookeeper-0.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc,zookeeper-test-zookeeper-1.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc,zookeeper-test-zookeeper-2.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc\",\n        \"zookeeper_auth_type\": \"none\",\n        \"zookeeper_port\": \"2181\"\n      }\n    }\n  kafka-env.sh.tmpl: \"export JAVA_OPTS=\\\"-Dsun.net.inetaddr.ttl=60 -Dsun.net.inetaddr.negative.ttl=60\n    ${JAVA_OPTS}\\\"\\nexport KAFKA_SERVER_MEMORY={{ getv \\\"/java_opts/memory_opts/kafka_memory\\\"\n    \\\"1024\\\" }}m\\n\\n{{- if eq (getv \\\"/security/auth_type\\\") \\\"kerberos\\\" }}\\nexport\n    JAVA_OPTS=\\\"-Djava.security.krb5.conf=/etc/krb5.conf \\n                        -Djava.security.auth.login.config=/etc/kafka/conf/jaas.conf\n    \\\\\\n                        -Dzookeeper.server.principal={{ getv \\\"/zookeeper_client_config/zookeeper_principal\\\"\n    \\\"\\\" }} \\\\\\n                        ${JAVA_OPTS}\\\"\\n{{- end }}\"\n  kafka.toml: |-\n    [[template]]\n    src = \"consumer.properties.tmpl\"\n    dest = \"/etc/kafka/conf/consumer.properties\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"producer.properties.tmpl\"\n    dest = \"/etc/kafka/conf/producer.properties\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"jaas.conf.tmpl\"\n    dest = \"/etc/kafka/conf/jaas.conf\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"kafka-env.sh.tmpl\"\n    dest = \"/etc/kafka/conf/kafka-env.sh\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"server.properties.tmpl\"\n    dest = \"/etc/kafka/conf/server.properties\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n  producer.properties.tmpl: |-\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" -}}\n    bootstrap.servers=${HOSTNAME}:9092\n    sasl.mechanism=GSSAPI\n    security.protocol=SASL_PLAINTEXT\n    sasl.kerberos.service.name={{ getv \"/security/guardian_principal_user\" \"kafka\" }}\n    sasl.kerberos.service.principal.instance={{ getv \"/security/guardian_principal_host\" \"tos\" }}\n    {{- else }}\n    bootstrap.servers=${HOSTNAME}:9092\n    security.protocol=PLAINTEXT\n    sasl.mechanism=PLAIN\n    {{- end }}\n  server.properties.tmpl: |\n    broker.id={{ getenv \"BROKER_ID\" }}\n\n    {{- range gets \"/server_properties/*\" }}\n    {{base .Key}}={{.Value}}\n    {{- end }}\n\n    {{- $KAFKA_ZK_ADDRESS := split (getv \"/zookeeper_client_config/zookeeper_addresses\" \"\") \",\" }}\n    zookeeper.connect={{join $KAFKA_ZK_ADDRESS (printf \":%s,\" (getv \"/zookeeper_client_config/zookeeper_port\" \"2181\"))}}:{{(getv \"/zookeeper_client_config/zookeeper_port\" \"2181\")}}\n\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    listeners=SASL_PLAINTEXT://0.0.0.0:9092\n    advertised.listeners=SASL_PLAINTEXT://{{getenv \"KAFKA_HOSTNAME\" \"localhost\"}}:9092\n    security.inter.broker.protocol=SASL_PLAINTEXT\n    sasl.mechanism.inter.broker.protocol=GSSAPI\n    sasl.enabled.mechanisms=GSSAPI\n\n    sasl.kerberos.service.name={{ getv \"/security/guardian_principal_user\" \"kafka\" }}\n    authorizer.class.name=io.transwarp.guardian.plugins.kafka.GuardianAclAuthorizer\n    super.users=User:{{ getv \"/security/guardian_principal_user\" \"tos\" }}\n    zookeeper.set.acl=true\n    sasl.kerberos.service.principal.instance={{ getv \"/security/guardian_principal_host\" \"kafka\" }}\n    sasl.kerberos.principal.to.local.rules=RULE:[1:$1@$0](^.*@.*$)s/^(.*)@.*$/$1/g,RULE:[2:$1@$0](^.*@.*$)s/^(.*)@.*$/$1/g,DEFAULT\n    {{- else }}\n    listeners=PLAINTEXT://0.0.0.0:9092\n    advertised.listeners=PLAINTEXT://{{getenv \"KAFKA_HOSTNAME\" \"localhost\"}}:9092\n    security.inter.broker.protocol=PLAINTEXT\n    sasl.mechanism.inter.broker.protocol=PLAIN\n    sasl.enabled.mechanisms=PLAIN\n    {{- end }}\n  tdh-env.sh.tmpl: |\n    #!/bin/bash\n    set -x\n\n    setup_keytab() {\n      echo \"setup_keytab\"\n    {{ if eq (getv \"/security/auth_type\") \"kerberos\" }}\n      # link_keytab\n      export KRB_MOUNTED_CONF_PATH=${KRB_MOUNTED_CONF_PATH:-/var/run/secrets/transwarp.io/tosvolume/keytab/krb5.conf}\n      export KRB_MOUNTED_KEYTAB=${KRB_MOUNTED_KEYTAB:-/var/run/secrets/transwarp.io/tosvolume/keytab/keytab}\n      if [ ! -f $KRB_MOUNTED_CONF_PATH ]; then\n        echo \"Expect krb5.conf at $KRB_MOUNTED_CONF_PATH but not found!\"\n        exit 1\n      fi\n      if [ ! -f $KRB_MOUNTED_KEYTAB ]; then\n        echo \"Expect keytab file at $KRB_MOUNTED_KEYTAB but not found!\"\n        exit 1\n      fi\n      ln -svf $KRB_MOUNTED_CONF_PATH /etc/krb5.conf\n      [ -d /etc/keytabs ] || mkdir -p /etc/keytabs\n      ln -svf $KRB_MOUNTED_KEYTAB /etc/keytabs/keytab\n    {{ end }}\n    }\n  tdh-env.toml: |-\n    [[template]]\n    src = \"tdh-env.sh.tmpl\"\n    dest = \"/etc/tdh-env.sh\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\nkind: ConfigMap\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: kafka\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: kafka-confd-conf\n    app.kubernetes.io/part-of: kafka\n    app.kubernetes.io/version: \"6.1\"\n  name: kafka-test-kafka-confd-conf\n  namespace: helmreleasetest-fixedns1\n\n---\napiVersion: v1\ndata:\n  entrypoint.sh: |\n    #!/bin/bash\n    set -x\n\n    export KAFKA_LOG_DIRS=/var/log/kafka\n    export KAFKA_CONF_DIR=/etc/kafka/conf\n    export KAFKA_HOSTNAME=$(echo \"`hostname -f`\" | sed -e 's/^[ \\t]*//g' -e 's/[ \\t]*$//g')\n\n    export BROKER_ID=${HOSTNAME##*-}\n\n    JMX_PORT=${JMX_PORT:-9999}\n\n    mkdir -p ${KAFKA_CONF_DIR} ${KAFKA_LOG_DIRS} /data\n    chown kafka:kafka ${KAFKA_CONF_DIR} ${KAFKA_LOG_DIRS} /data\n\n    confd -onetime -backend file -prefix / -file /etc/confd/kafka-confd.conf\n\n    [ -s /etc/guardian-site.xml ] && {\n      cp /etc/guardian-site.xml $KAFKA_CONF_DIR\n    }\n\n    KAFKA_ENV=$KAFKA_CONF_DIR/kafka-env.sh\n\n    JMXEXPORTER_ENABLED=${JMXEXPORTER_ENABLED:-\"false\"}\n    JMX_EXPORTER_JAR=`ls /usr/lib/jmx_exporter/jmx_prometheus_javaagent-*.jar | head -n1`\n    [ \"${JMXEXPORTER_ENABLED}\" = \"true\" ] && \\\n      export JMXEXPORTER_OPTS=\" -javaagent:${JMX_EXPORTER_JAR}=${JMXEXPORTER_PORT:-\"19009\"}:/usr/lib/jmx_exporter/configs/kafka.yml \"\n\n    export KAFKA_JMX_OPTS=\"-Dcom.sun.management.jmxremote=true -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.rmi.port=$JMX_PORT\\\n     -Dcom.sun.management.jmxremote.port=$JMX_PORT\"\n\n\n    [ -f $KAFKA_ENV ] && {\n      source $KAFKA_ENV\n    }\n    [ -f /etc/tdh-env.sh ] && {\n      source /etc/tdh-env.sh\n      setup_keytab\n\n    }\n\n    CLASSPATH=\"\"\n    set +x\n    for jar in `find /usr/lib/kafka -name \"*.jar\"`\n    do\n       CLASSPATH+=\":$jar\"\n    done\n    for jar in `find /usr/lib/guardian-plugins/lib -name \"*.jar\"`\n    do\n       CLASSPATH+=\":$jar\"\n    done\n    CLASSPATH+=\":${KAFKA_CONF_DIR}\"\n    set -x\n\n    JAVA_OPTS=\"-Xmx${KAFKA_SERVER_MEMORY} \\\n    -Xms${KAFKA_SERVER_MEMORY} \\\n    -XX:+UseCompressedOops \\\n    -XX:+UseParNewGC \\\n    -XX:+UseConcMarkSweepGC \\\n    -XX:+CMSClassUnloadingEnabled \\\n    -XX:+CMSScavengeBeforeRemark \\\n    -XX:+DisableExplicitGC \\\n    -Djava.awt.headless=true \\\n    -Xloggc:${KAFKA_LOG_DIRS}/kafkaServer-gc.log \\\n    -verbose:gc \\\n    -XX:+PrintGCDetails \\\n    -XX:+PrintGCDateStamps \\\n    -XX:+PrintGCTimeStamps \\\n    -Dlog4j.configuration=file:/etc/kafka/conf/log4j.properties \\\n    -Dkafka.logs.dir=${KAFKA_LOG_DIRS}/kafkaServer.log \\\n    $JMXEXPORTER_OPTS $KAFKA_JMX_OPTS $JAVA_OPTS\"\n\n    $JAVA_HOME/bin/java $JAVA_OPTS -cp $CLASSPATH kafka.Kafka /etc/kafka/conf/server.properties\nkind: ConfigMap\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: kafka\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: kafka-entrypoint\n    app.kubernetes.io/part-of: kafka\n    app.kubernetes.io/version: \"6.1\"\n  name: kafka-test-kafka-entrypoint\n  namespace: helmreleasetest-fixedns1\n\n---\napiVersion: v1\nkind: Service\nmetadata:\n  annotations:\n    service.alpha.kubernetes.io/tolerate-unready-endpoints: \"true\"\n  labels:\n    app.kubernetes.io/component: kafka\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: kafka\n    app.kubernetes.io/part-of: kafka\n    app.kubernetes.io/service-type: headless-service\n    app.kubernetes.io/version: \"6.1\"\n  name: kafka-test-kafka-hl\n  namespace: helmreleasetest-fixedns1\nspec:\n  clusterIP: None\n  ports:\n  - name: web\n    port: 9092\n    targetPort: 9092\n  selector:\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/name: kafka\n    app.kubernetes.io/version: \"6.1\"\n\n---\napiVersion: v1\nkind: Service\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: kafka\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: kafka\n    app.kubernetes.io/part-of: kafka\n    app.kubernetes.io/service-type: nodeport-service\n    app.kubernetes.io/version: \"6.1\"\n  name: kafka-test-kafka\n  namespace: helmreleasetest-fixedns1\nspec:\n  ports:\n  - name: web\n    port: 9092\n    targetPort: 9092\n  selector:\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/name: kafka\n    app.kubernetes.io/version: \"6.1\"\n  type: NodePort\n\n---\napiVersion: apps/v1beta1\nkind: StatefulSet\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: kafka\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: kafka\n    app.kubernetes.io/part-of: kafka\n    app.kubernetes.io/version: \"6.1\"\n  name: kafka-test-kafka\n  namespace: helmreleasetest-fixedns1\nspec:\n  podManagementPolicy: Parallel\n  replicas: 3\n  selector:\n    matchLabels:\n      app.kubernetes.io/instance: kafka-test\n      app.kubernetes.io/name: kafka\n      app.kubernetes.io/version: \"6.1\"\n  serviceName: kafka-test-kafka-hl\n  template:\n    metadata:\n      annotations:\n        tos.network.staticIP: \"true\"\n        transwarp/configmap.md5: 9969e312241e3505aa907ed898c8a9ed\n      labels:\n        app.kubernetes.io/instance: kafka-test\n        app.kubernetes.io/name: kafka\n        app.kubernetes.io/version: \"6.1\"\n    spec:\n      affinity:\n        podAntiAffinity:\n          requiredDuringSchedulingIgnoredDuringExecution:\n          - labelSelector:\n              matchLabels:\n                app.kubernetes.io/instance: kafka-test\n                app.kubernetes.io/name: kafka\n                app.kubernetes.io/version: \"6.1\"\n            namespaces:\n            - helmreleasetest-fixedns1\n            topologyKey: kubernetes.io/hostname\n      containers:\n      - command:\n        - /boot/entrypoint.sh\n        env: []\n        image: docker.io/corndai1997/kafka:6.0\n        imagePullPolicy: Always\n        name: kafka\n        resources:\n          limits:\n            cpu: \"0.20000000000000001\"\n            memory: 200Mi\n            nvidia.com/gpu: \"0\"\n          requests:\n            cpu: \"0.10000000000000001\"\n            memory: 100Mi\n            nvidia.com/gpu: \"0\"\n        volumeMounts:\n        - mountPath: /boot\n          name: kafka-entrypoint\n        - mountPath: /etc/confd\n          name: kafka-confd-conf\n        - mountPath: /data\n          name: data\n        - mountPath: /var/log/kafka\n          name: log\n      hostNetwork: false\n      initContainers: []\n      priorityClassName: low-priority\n      restartPolicy: Always\n      terminationGracePeriodSeconds: 30\n      volumes:\n      - configMap:\n          items:\n          - key: entrypoint.sh\n            mode: 493\n            path: entrypoint.sh\n          name: kafka-test-kafka-entrypoint\n        name: kafka-entrypoint\n      - configMap:\n          items:\n          - key: kafka-confd.conf\n            path: kafka-confd.conf\n          - key: kafka.toml\n            path: conf.d/kafka.toml\n          - key: tdh-env.toml\n            path: conf.d/tdh-env.toml\n          - key: jaas.conf.tmpl\n            path: templates/jaas.conf.tmpl\n          - key: kafka-env.sh.tmpl\n            path: templates/kafka-env.sh.tmpl\n          - key: server.properties.tmpl\n            path: templates/server.properties.tmpl\n          - key: producer.properties.tmpl\n            path: templates/producer.properties.tmpl\n          - key: consumer.properties.tmpl\n            path: templates/consumer.properties.tmpl\n          - key: tdh-env.sh.tmpl\n            path: templates/tdh-env.sh.tmpl\n          name: kafka-test-kafka-confd-conf\n        name: kafka-confd-conf\n      - name: log\n        tosDisk:\n          accessMode: ReadWriteOnce\n          capability: 100Gi\n          name: log\n          storageType: local\n  updateStrategy:\n    type: RollingUpdate\n  volumeClaimTemplates:\n  - metadata:\n      annotations:\n        volume.beta.kubernetes.io/storage-class: local\n      labels:\n        app.kubernetes.io/component: kafka\n        app.kubernetes.io/instance: kafka-test\n        app.kubernetes.io/managed-by: walm\n        app.kubernetes.io/name: kafka\n        app.kubernetes.io/part-of: kafka\n        app.kubernetes.io/version: \"6.1\"\n      name: data\n    spec:\n      accessModes:\n      - ReadWriteOnce\n      resources:\n        requests:\n          storage: 100Gi\n      storageClassName: local\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: kafka-test\n  namespace: helmreleasetest-fixedns1\nspec:\n  chartAppVersion: \"6.1\"\n  chartImage: \"\"\n  chartName: kafka\n  chartVersion: 6.1.0\n  chartWalmVersion: v2\n  configValues: {}\n  dependencies:\n    zookeeper: zookeeper-test\n  dependenciesConfigValues:\n    ZOOKEEPER_CLIENT_CONFIG:\n      zookeeper_addresses: zookeeper-test-zookeeper-0.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc,zookeeper-test-zookeeper-1.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc,zookeeper-test-zookeeper-2.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc\n      zookeeper_auth_type: none\n      zookeeper_port: \"2181\"\n  isomateConfig: null\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n"
				Expect(releaseCache.Manifest).To(Equal(manifest))

				By("install kafka which depends on zookeeper in other namespace")
				releaseRequest = &release.ReleaseRequestV2{
					ReleaseRequest: release.ReleaseRequest{
						Name:         "kafka-test",
						Dependencies: map[string]string{"zookeeper": fixedNamespace1 + "/zookeeper-test"},
					},
				}

				releaseCache, err = helm.InstallOrCreateRelease(fixedNamespace2, releaseRequest, chartFiles, false, false, nil)
				Expect(err).NotTo(HaveOccurred())

				assertReleaseCacheBasic(releaseCache, fixedNamespace2, "kafka-test", "", "kafka",
					"6.1.0", "6.1", 1)

				defaultValues, err = getChartDefaultValues(chartFiles)
				Expect(err).NotTo(HaveOccurred())
				computedValues = map[string]interface{}{}
				computedValues = util.MergeValues(computedValues, defaultValues, false)
				computedValues = util.MergeValues(computedValues, map[string]interface{}{
					plugins.WalmPluginConfigKey: []*k8s.ReleasePlugin{
						{
							Name: plugins.ValidateReleaseConfigPluginName,
						},
						{
							Name: plugins.IsomateSetConverterPluginName,
						},
					},
					"ZOOKEEPER_CLIENT_CONFIG": map[string]interface{}{
						"zookeeper_addresses": "zookeeper-test-zookeeper-0.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc,zookeeper-test-zookeeper-1.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc,zookeeper-test-zookeeper-2.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc",
						"zookeeper_auth_type": "none",
						"zookeeper_port":      "2181",
					},
				}, false)

				Expect(releaseCache.ComputedValues).To(Equal(computedValues))

				manifest = "\n---\napiVersion: v1\ndata:\n  consumer.properties.tmpl: |-\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" -}}\n    bootstrap.servers={{getenv \"KAFKA_HOSTNAME\" \"localhost\"}}:9092\n    sasl.mechanism=GSSAPI\n    security.protocol=SASL_PLAINTEXT\n    sasl.kerberos.service.name={{ getv \"/security/guardian_principal_user\" \"kafka\" }}\n    sasl.kerberos.service.principal.instance={{ getv \"/security/guardian_principal_host\" \"tos\" }}\n    {{- else }}\n    bootstrap.servers={{getenv \"KAFKA_HOSTNAME\" \"localhost\"}}:9092\n    security.protocol=PLAINTEXT\n    sasl.mechanism=PLAIN\n    {{- end }}\n  jaas.conf.tmpl: |\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    KafkaServer {\n      com.sun.security.auth.module.Krb5LoginModule required\n      useKeyTab=true\n      keyTab=\"/etc/keytabs/keytab\"\n      storeKey=true\n      useTicketCache=false\n      principal=\"{{ getv \"/security/guardian_principal_user\" \"kafka\" }}/{{ getv \"/security/guardian_principal_host\" \"tos\" }}@{{ getv \"/security/guardian_client_config/realm\" \"TDH\" }}\";\n    };\n    KafkaClient {\n      com.sun.security.auth.module.Krb5LoginModule required\n      useKeyTab=true\n      keyTab=\"/etc/keytabs/keytab\"\n      storeKey=true\n      useTicketCache=false\n      principal=\"{{ getv \"/security/guardian_principal_user\" \"kafka\" }}/{{ getv \"/security/guardian_principal_host\" \"tos\" }}@{{ getv \"/security/guardian_client_config/realm\" \"TDH\" }}\";\n    };\n    // Zookeeper client authentication\n    Client {\n      com.sun.security.auth.module.Krb5LoginModule required\n      useKeyTab=true\n      storeKey=true\n      useTicketCache=false\n      keyTab=\"/etc/keytabs/keytab\"\n      principal=\"{{ getv \"/security/guardian_principal_user\" \"kafka\" }}/{{ getv \"/security/guardian_principal_host\" \"tos\" }}@{{ getv \"/security/guardian_client_config/realm\" \"TDH\" }}\";\n    };\n    {{- end }}\n  kafka-confd.conf: |-\n    {\n      \"java_opts\": {\n        \"memory_opts\": {\n          \"kafka_memory\": \"3276\"\n        }\n      },\n      \"kafka\": {\n\n      },\n      \"security\": {\n        \"auth_type\": \"none\"\n      },\n      \"server_properties\": {\n        \"default.replication.factor\": 2,\n        \"log.dirs\": \"/data\",\n        \"log.flush.interval.messages\": 10000,\n        \"log.flush.interval.ms\": 1000,\n        \"log.retention.bytes\": 1073741824,\n        \"log.retention.check.interval.ms\": 300000,\n        \"log.retention.hours\": 6,\n        \"log.segment.bytes\": 1073741824,\n        \"message.max.bytes\": 100000000,\n        \"num.io.threads\": 8,\n        \"num.network.threads\": 3,\n        \"num.partitions\": 3,\n        \"num.recovery.threads.per.data.dir\": 1,\n        \"replica.fetch.max.bytes\": 100000000,\n        \"socket.receive.buffer.bytes\": 102400,\n        \"socket.request.max.bytes\": 104857600,\n        \"socket.send.buffer.bytes\": 102400,\n        \"zookeeper.connection.timeout.ms\": 6000\n      },\n      \"transwarpApplicationPause\": false,\n      \"transwarpCniNetwork\": \"overlay\",\n      \"transwarpGlobalIngress\": {\n        \"httpPort\": 80,\n        \"httpsPort\": 443\n      },\n      \"transwarpLicenseAddress\": \"\",\n      \"transwarpMetrics\": {\n        \"enable\": true\n      },\n      \"zookeeper_client_config\": {\n        \"zookeeper_addresses\": \"zookeeper-test-zookeeper-0.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc,zookeeper-test-zookeeper-1.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc,zookeeper-test-zookeeper-2.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc\",\n        \"zookeeper_auth_type\": \"none\",\n        \"zookeeper_port\": \"2181\"\n      }\n    }\n  kafka-env.sh.tmpl: \"export JAVA_OPTS=\\\"-Dsun.net.inetaddr.ttl=60 -Dsun.net.inetaddr.negative.ttl=60\n    ${JAVA_OPTS}\\\"\\nexport KAFKA_SERVER_MEMORY={{ getv \\\"/java_opts/memory_opts/kafka_memory\\\"\n    \\\"1024\\\" }}m\\n\\n{{- if eq (getv \\\"/security/auth_type\\\") \\\"kerberos\\\" }}\\nexport\n    JAVA_OPTS=\\\"-Djava.security.krb5.conf=/etc/krb5.conf \\n                        -Djava.security.auth.login.config=/etc/kafka/conf/jaas.conf\n    \\\\\\n                        -Dzookeeper.server.principal={{ getv \\\"/zookeeper_client_config/zookeeper_principal\\\"\n    \\\"\\\" }} \\\\\\n                        ${JAVA_OPTS}\\\"\\n{{- end }}\"\n  kafka.toml: |-\n    [[template]]\n    src = \"consumer.properties.tmpl\"\n    dest = \"/etc/kafka/conf/consumer.properties\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"producer.properties.tmpl\"\n    dest = \"/etc/kafka/conf/producer.properties\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"jaas.conf.tmpl\"\n    dest = \"/etc/kafka/conf/jaas.conf\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"kafka-env.sh.tmpl\"\n    dest = \"/etc/kafka/conf/kafka-env.sh\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"server.properties.tmpl\"\n    dest = \"/etc/kafka/conf/server.properties\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n  producer.properties.tmpl: |-\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" -}}\n    bootstrap.servers=${HOSTNAME}:9092\n    sasl.mechanism=GSSAPI\n    security.protocol=SASL_PLAINTEXT\n    sasl.kerberos.service.name={{ getv \"/security/guardian_principal_user\" \"kafka\" }}\n    sasl.kerberos.service.principal.instance={{ getv \"/security/guardian_principal_host\" \"tos\" }}\n    {{- else }}\n    bootstrap.servers=${HOSTNAME}:9092\n    security.protocol=PLAINTEXT\n    sasl.mechanism=PLAIN\n    {{- end }}\n  server.properties.tmpl: |\n    broker.id={{ getenv \"BROKER_ID\" }}\n\n    {{- range gets \"/server_properties/*\" }}\n    {{base .Key}}={{.Value}}\n    {{- end }}\n\n    {{- $KAFKA_ZK_ADDRESS := split (getv \"/zookeeper_client_config/zookeeper_addresses\" \"\") \",\" }}\n    zookeeper.connect={{join $KAFKA_ZK_ADDRESS (printf \":%s,\" (getv \"/zookeeper_client_config/zookeeper_port\" \"2181\"))}}:{{(getv \"/zookeeper_client_config/zookeeper_port\" \"2181\")}}\n\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    listeners=SASL_PLAINTEXT://0.0.0.0:9092\n    advertised.listeners=SASL_PLAINTEXT://{{getenv \"KAFKA_HOSTNAME\" \"localhost\"}}:9092\n    security.inter.broker.protocol=SASL_PLAINTEXT\n    sasl.mechanism.inter.broker.protocol=GSSAPI\n    sasl.enabled.mechanisms=GSSAPI\n\n    sasl.kerberos.service.name={{ getv \"/security/guardian_principal_user\" \"kafka\" }}\n    authorizer.class.name=io.transwarp.guardian.plugins.kafka.GuardianAclAuthorizer\n    super.users=User:{{ getv \"/security/guardian_principal_user\" \"tos\" }}\n    zookeeper.set.acl=true\n    sasl.kerberos.service.principal.instance={{ getv \"/security/guardian_principal_host\" \"kafka\" }}\n    sasl.kerberos.principal.to.local.rules=RULE:[1:$1@$0](^.*@.*$)s/^(.*)@.*$/$1/g,RULE:[2:$1@$0](^.*@.*$)s/^(.*)@.*$/$1/g,DEFAULT\n    {{- else }}\n    listeners=PLAINTEXT://0.0.0.0:9092\n    advertised.listeners=PLAINTEXT://{{getenv \"KAFKA_HOSTNAME\" \"localhost\"}}:9092\n    security.inter.broker.protocol=PLAINTEXT\n    sasl.mechanism.inter.broker.protocol=PLAIN\n    sasl.enabled.mechanisms=PLAIN\n    {{- end }}\n  tdh-env.sh.tmpl: |\n    #!/bin/bash\n    set -x\n\n    setup_keytab() {\n      echo \"setup_keytab\"\n    {{ if eq (getv \"/security/auth_type\") \"kerberos\" }}\n      # link_keytab\n      export KRB_MOUNTED_CONF_PATH=${KRB_MOUNTED_CONF_PATH:-/var/run/secrets/transwarp.io/tosvolume/keytab/krb5.conf}\n      export KRB_MOUNTED_KEYTAB=${KRB_MOUNTED_KEYTAB:-/var/run/secrets/transwarp.io/tosvolume/keytab/keytab}\n      if [ ! -f $KRB_MOUNTED_CONF_PATH ]; then\n        echo \"Expect krb5.conf at $KRB_MOUNTED_CONF_PATH but not found!\"\n        exit 1\n      fi\n      if [ ! -f $KRB_MOUNTED_KEYTAB ]; then\n        echo \"Expect keytab file at $KRB_MOUNTED_KEYTAB but not found!\"\n        exit 1\n      fi\n      ln -svf $KRB_MOUNTED_CONF_PATH /etc/krb5.conf\n      [ -d /etc/keytabs ] || mkdir -p /etc/keytabs\n      ln -svf $KRB_MOUNTED_KEYTAB /etc/keytabs/keytab\n    {{ end }}\n    }\n  tdh-env.toml: |-\n    [[template]]\n    src = \"tdh-env.sh.tmpl\"\n    dest = \"/etc/tdh-env.sh\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\nkind: ConfigMap\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: kafka\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: kafka-confd-conf\n    app.kubernetes.io/part-of: kafka\n    app.kubernetes.io/version: \"6.1\"\n  name: kafka-test-kafka-confd-conf\n  namespace: helmreleasetest-fixedns2\n\n---\napiVersion: v1\ndata:\n  entrypoint.sh: |\n    #!/bin/bash\n    set -x\n\n    export KAFKA_LOG_DIRS=/var/log/kafka\n    export KAFKA_CONF_DIR=/etc/kafka/conf\n    export KAFKA_HOSTNAME=$(echo \"`hostname -f`\" | sed -e 's/^[ \\t]*//g' -e 's/[ \\t]*$//g')\n\n    export BROKER_ID=${HOSTNAME##*-}\n\n    JMX_PORT=${JMX_PORT:-9999}\n\n    mkdir -p ${KAFKA_CONF_DIR} ${KAFKA_LOG_DIRS} /data\n    chown kafka:kafka ${KAFKA_CONF_DIR} ${KAFKA_LOG_DIRS} /data\n\n    confd -onetime -backend file -prefix / -file /etc/confd/kafka-confd.conf\n\n    [ -s /etc/guardian-site.xml ] && {\n      cp /etc/guardian-site.xml $KAFKA_CONF_DIR\n    }\n\n    KAFKA_ENV=$KAFKA_CONF_DIR/kafka-env.sh\n\n    JMXEXPORTER_ENABLED=${JMXEXPORTER_ENABLED:-\"false\"}\n    JMX_EXPORTER_JAR=`ls /usr/lib/jmx_exporter/jmx_prometheus_javaagent-*.jar | head -n1`\n    [ \"${JMXEXPORTER_ENABLED}\" = \"true\" ] && \\\n      export JMXEXPORTER_OPTS=\" -javaagent:${JMX_EXPORTER_JAR}=${JMXEXPORTER_PORT:-\"19009\"}:/usr/lib/jmx_exporter/configs/kafka.yml \"\n\n    export KAFKA_JMX_OPTS=\"-Dcom.sun.management.jmxremote=true -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote.rmi.port=$JMX_PORT\\\n     -Dcom.sun.management.jmxremote.port=$JMX_PORT\"\n\n\n    [ -f $KAFKA_ENV ] && {\n      source $KAFKA_ENV\n    }\n    [ -f /etc/tdh-env.sh ] && {\n      source /etc/tdh-env.sh\n      setup_keytab\n\n    }\n\n    CLASSPATH=\"\"\n    set +x\n    for jar in `find /usr/lib/kafka -name \"*.jar\"`\n    do\n       CLASSPATH+=\":$jar\"\n    done\n    for jar in `find /usr/lib/guardian-plugins/lib -name \"*.jar\"`\n    do\n       CLASSPATH+=\":$jar\"\n    done\n    CLASSPATH+=\":${KAFKA_CONF_DIR}\"\n    set -x\n\n    JAVA_OPTS=\"-Xmx${KAFKA_SERVER_MEMORY} \\\n    -Xms${KAFKA_SERVER_MEMORY} \\\n    -XX:+UseCompressedOops \\\n    -XX:+UseParNewGC \\\n    -XX:+UseConcMarkSweepGC \\\n    -XX:+CMSClassUnloadingEnabled \\\n    -XX:+CMSScavengeBeforeRemark \\\n    -XX:+DisableExplicitGC \\\n    -Djava.awt.headless=true \\\n    -Xloggc:${KAFKA_LOG_DIRS}/kafkaServer-gc.log \\\n    -verbose:gc \\\n    -XX:+PrintGCDetails \\\n    -XX:+PrintGCDateStamps \\\n    -XX:+PrintGCTimeStamps \\\n    -Dlog4j.configuration=file:/etc/kafka/conf/log4j.properties \\\n    -Dkafka.logs.dir=${KAFKA_LOG_DIRS}/kafkaServer.log \\\n    $JMXEXPORTER_OPTS $KAFKA_JMX_OPTS $JAVA_OPTS\"\n\n    $JAVA_HOME/bin/java $JAVA_OPTS -cp $CLASSPATH kafka.Kafka /etc/kafka/conf/server.properties\nkind: ConfigMap\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: kafka\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: kafka-entrypoint\n    app.kubernetes.io/part-of: kafka\n    app.kubernetes.io/version: \"6.1\"\n  name: kafka-test-kafka-entrypoint\n  namespace: helmreleasetest-fixedns2\n\n---\napiVersion: v1\nkind: Service\nmetadata:\n  annotations:\n    service.alpha.kubernetes.io/tolerate-unready-endpoints: \"true\"\n  labels:\n    app.kubernetes.io/component: kafka\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: kafka\n    app.kubernetes.io/part-of: kafka\n    app.kubernetes.io/service-type: headless-service\n    app.kubernetes.io/version: \"6.1\"\n  name: kafka-test-kafka-hl\n  namespace: helmreleasetest-fixedns2\nspec:\n  clusterIP: None\n  ports:\n  - name: web\n    port: 9092\n    targetPort: 9092\n  selector:\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/name: kafka\n    app.kubernetes.io/version: \"6.1\"\n\n---\napiVersion: v1\nkind: Service\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: kafka\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: kafka\n    app.kubernetes.io/part-of: kafka\n    app.kubernetes.io/service-type: nodeport-service\n    app.kubernetes.io/version: \"6.1\"\n  name: kafka-test-kafka\n  namespace: helmreleasetest-fixedns2\nspec:\n  ports:\n  - name: web\n    port: 9092\n    targetPort: 9092\n  selector:\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/name: kafka\n    app.kubernetes.io/version: \"6.1\"\n  type: NodePort\n\n---\napiVersion: apps/v1beta1\nkind: StatefulSet\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: kafka\n    app.kubernetes.io/instance: kafka-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: kafka\n    app.kubernetes.io/part-of: kafka\n    app.kubernetes.io/version: \"6.1\"\n  name: kafka-test-kafka\n  namespace: helmreleasetest-fixedns2\nspec:\n  podManagementPolicy: Parallel\n  replicas: 3\n  selector:\n    matchLabels:\n      app.kubernetes.io/instance: kafka-test\n      app.kubernetes.io/name: kafka\n      app.kubernetes.io/version: \"6.1\"\n  serviceName: kafka-test-kafka-hl\n  template:\n    metadata:\n      annotations:\n        tos.network.staticIP: \"true\"\n        transwarp/configmap.md5: 9969e312241e3505aa907ed898c8a9ed\n      labels:\n        app.kubernetes.io/instance: kafka-test\n        app.kubernetes.io/name: kafka\n        app.kubernetes.io/version: \"6.1\"\n    spec:\n      affinity:\n        podAntiAffinity:\n          requiredDuringSchedulingIgnoredDuringExecution:\n          - labelSelector:\n              matchLabels:\n                app.kubernetes.io/instance: kafka-test\n                app.kubernetes.io/name: kafka\n                app.kubernetes.io/version: \"6.1\"\n            namespaces:\n            - helmreleasetest-fixedns2\n            topologyKey: kubernetes.io/hostname\n      containers:\n      - command:\n        - /boot/entrypoint.sh\n        env: []\n        image: docker.io/corndai1997/kafka:6.0\n        imagePullPolicy: Always\n        name: kafka\n        resources:\n          limits:\n            cpu: \"0.20000000000000001\"\n            memory: 200Mi\n            nvidia.com/gpu: \"0\"\n          requests:\n            cpu: \"0.10000000000000001\"\n            memory: 100Mi\n            nvidia.com/gpu: \"0\"\n        volumeMounts:\n        - mountPath: /boot\n          name: kafka-entrypoint\n        - mountPath: /etc/confd\n          name: kafka-confd-conf\n        - mountPath: /data\n          name: data\n        - mountPath: /var/log/kafka\n          name: log\n      hostNetwork: false\n      initContainers: []\n      priorityClassName: low-priority\n      restartPolicy: Always\n      terminationGracePeriodSeconds: 30\n      volumes:\n      - configMap:\n          items:\n          - key: entrypoint.sh\n            mode: 493\n            path: entrypoint.sh\n          name: kafka-test-kafka-entrypoint\n        name: kafka-entrypoint\n      - configMap:\n          items:\n          - key: kafka-confd.conf\n            path: kafka-confd.conf\n          - key: kafka.toml\n            path: conf.d/kafka.toml\n          - key: tdh-env.toml\n            path: conf.d/tdh-env.toml\n          - key: jaas.conf.tmpl\n            path: templates/jaas.conf.tmpl\n          - key: kafka-env.sh.tmpl\n            path: templates/kafka-env.sh.tmpl\n          - key: server.properties.tmpl\n            path: templates/server.properties.tmpl\n          - key: producer.properties.tmpl\n            path: templates/producer.properties.tmpl\n          - key: consumer.properties.tmpl\n            path: templates/consumer.properties.tmpl\n          - key: tdh-env.sh.tmpl\n            path: templates/tdh-env.sh.tmpl\n          name: kafka-test-kafka-confd-conf\n        name: kafka-confd-conf\n      - name: log\n        tosDisk:\n          accessMode: ReadWriteOnce\n          capability: 100Gi\n          name: log\n          storageType: local\n  updateStrategy:\n    type: RollingUpdate\n  volumeClaimTemplates:\n  - metadata:\n      annotations:\n        volume.beta.kubernetes.io/storage-class: local\n      labels:\n        app.kubernetes.io/component: kafka\n        app.kubernetes.io/instance: kafka-test\n        app.kubernetes.io/managed-by: walm\n        app.kubernetes.io/name: kafka\n        app.kubernetes.io/part-of: kafka\n        app.kubernetes.io/version: \"6.1\"\n      name: data\n    spec:\n      accessModes:\n      - ReadWriteOnce\n      resources:\n        requests:\n          storage: 100Gi\n      storageClassName: local\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: kafka-test\n  namespace: helmreleasetest-fixedns2\nspec:\n  chartAppVersion: \"6.1\"\n  chartImage: \"\"\n  chartName: kafka\n  chartVersion: 6.1.0\n  chartWalmVersion: v2\n  configValues: {}\n  dependencies:\n    zookeeper: helmreleasetest-fixedns1/zookeeper-test\n  dependenciesConfigValues:\n    ZOOKEEPER_CLIENT_CONFIG:\n      zookeeper_addresses: zookeeper-test-zookeeper-0.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc,zookeeper-test-zookeeper-1.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc,zookeeper-test-zookeeper-2.zookeeper-test-zookeeper-hl.helmreleasetest-fixedns1.svc\n      zookeeper_auth_type: none\n      zookeeper_port: \"2181\"\n  isomateConfig: null\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n"
				Expect(releaseCache.Manifest).To(Equal(manifest))
			})
		})

		It("test release update", func() {
			By("update release with local chart")
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "tomcat-test",
				},
			}

			_, err := helm.InstallOrCreateRelease(namespace, releaseRequest, tomcatChartFiles, false, false, nil)
			Expect(err).NotTo(HaveOccurred())

			releaseRequest.ConfigValues = map[string]interface{}{
				"replicaCount": 2,
			}
			releaseInfo := &release.ReleaseInfoV2{}

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, tomcatChartFiles, false, true, releaseInfo)
			Expect(err).NotTo(HaveOccurred())

			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 2)

			tomcatComputedValues["replicaCount"] = 2
			assertYamlConfigValues(releaseCache.ComputedValues, tomcatComputedValues)

			manifest := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 2\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartImage: \"\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  chartWalmVersion: v2\n  configValues:\n    replicaCount: 2\n  dependencies: {}\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n",
				namespace, namespace, namespace)
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

			mataInfoParams := getTomcatChartDefaultMetaInfoParams()
			replicas := int64(2)
			mataInfoParams.Roles[1].RoleBaseConfigValue.Replicas = &replicas
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))

			By("update release will reuse the previous config values ")
			releaseRequest.ConfigValues = nil
			releaseInfo.ConfigValues = map[string]interface{}{
				"replicaCount": 2,
			}
			releaseCache, err = helm.InstallOrCreateRelease(namespace, releaseRequest, tomcatChartFiles, false, true, releaseInfo)
			Expect(err).NotTo(HaveOccurred())

			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 3)

			assertYamlConfigValues(releaseCache.ComputedValues, tomcatComputedValues)
			Expect(releaseCache.Manifest).To(Equal(manifest))
			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})

		It("test release pause", func() {
			By("install release")
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "tomcat-test",
				},
			}

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, tomcatChartFiles, false, false, nil)
			Expect(err).NotTo(HaveOccurred())

			By("pause release")
			oldReleaseInfo := &release.ReleaseInfoV2{}
			oldReleaseInfo.Namespace = releaseCache.Namespace
			oldReleaseInfo.Name = releaseCache.Name

			releaseCache, err = helm.PauseOrRecoverRelease(true, oldReleaseInfo)
			Expect(err).NotTo(HaveOccurred())

			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 2)

			tomcatComputedValues = util.MergeValues(tomcatComputedValues, map[string]interface{}{
				plugins.WalmPluginConfigKey: []*k8s.ReleasePlugin{
					{
						Version: "1.0",
						Name:    plugins.PauseReleasePluginName,
					},
					{
						Name: plugins.ValidateReleaseConfigPluginName,
					},
					{
						Name: plugins.IsomateSetConverterPluginName,
					},
				},
			}, false)
			assertYamlConfigValues(releaseCache.ComputedValues, tomcatComputedValues)

			manifest := strings.ReplaceAll("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: helmreleasetest-dqrhw\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: helmreleasetest-dqrhw\nspec:\n  replicas: 0\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: helmreleasetest-dqrhw\nspec:\n  chartAppVersion: \"7\"\n  chartImage: \"\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  chartWalmVersion: v2\n  configValues: {}\n  dependencies: null\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n",
				"helmreleasetest-dqrhw", namespace)
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

			mataInfoParams := getTomcatChartDefaultMetaInfoParams()
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))

			By("recover release")
			oldReleaseInfo.Plugins = []*k8s.ReleasePlugin{
				{
					Version: "1.0",
					Name:    plugins.PauseReleasePluginName,
				},
			}
			releaseCache, err = helm.PauseOrRecoverRelease(false, oldReleaseInfo)
			Expect(err).NotTo(HaveOccurred())

			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 3)

			tomcatComputedValues = util.MergeValues(tomcatComputedValues, map[string]interface{}{
				plugins.WalmPluginConfigKey: []*k8s.ReleasePlugin{
					{
						Version: "1.0",
						Name:    plugins.PauseReleasePluginName,
						Disable: true,
					},
					{
						Name: plugins.ValidateReleaseConfigPluginName,
					},
					{
						Name: plugins.IsomateSetConverterPluginName,
					},
				},
			}, false)
			assertYamlConfigValues(releaseCache.ComputedValues, tomcatComputedValues)

			manifest = fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartImage: \"\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  chartWalmVersion: v2\n  configValues: {}\n  dependencies: null\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n",
				namespace, namespace, namespace)
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

			mataInfoParams = getTomcatChartDefaultMetaInfoParams()
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})

		Describe("need another random namespace", func() {
			var (
				anotherNamespace string
			)

			BeforeEach(func() {
				By("create another namespace")
				anotherNamespace, err = framework.CreateRandomNamespace("helmReleaseTest", nil)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				By("delete another namespace")
				err = framework.DeleteNamespace(anotherNamespace)
				Expect(err).NotTo(HaveOccurred())
			})
			It("test list & delete release", func() {
				By("list releases")
				releaseRequest := &release.ReleaseRequestV2{
					ReleaseRequest: release.ReleaseRequest{
						Name: "tomcat-test",
					},
				}

				_, err := helm.InstallOrCreateRelease(namespace, releaseRequest, tomcatChartFiles, false, false, nil)
				Expect(err).NotTo(HaveOccurred())

				_, err = helm.InstallOrCreateRelease(anotherNamespace, releaseRequest, tomcatChartFiles, false, false, nil)
				Expect(err).NotTo(HaveOccurred())

				releaseCaches, err := helm.ListAllReleases()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(releaseCaches) >= 2).To(BeTrue())

				getReleaseCache := func(releaseCaches []*release.ReleaseCache, namespace, name string) *release.ReleaseCache {
					for _, releaseCache := range releaseCaches {
						if releaseCache.Name == name && releaseCache.Namespace == namespace {
							return releaseCache
						}
					}
					return nil
				}
				releaseCache1 := getReleaseCache(releaseCaches, namespace, "tomcat-test")
				Expect(releaseCache1).NotTo(BeNil())
				releaseCache2 := getReleaseCache(releaseCaches, anotherNamespace, "tomcat-test")
				Expect(releaseCache2).NotTo(BeNil())

				assertReleaseCacheBasic(releaseCache1, namespace, "tomcat-test", "", "tomcat",
					"0.2.0", "7", 1)

				assertYamlConfigValues(releaseCache1.ComputedValues, tomcatComputedValues)

				manifest := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartImage: \"\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  chartWalmVersion: v2\n  configValues: {}\n  dependencies: null\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n",
					namespace, namespace, namespace)
				Expect(releaseCache1.Manifest).To(Equal(manifest))

				Expect(releaseCache1.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

				mataInfoParams := getTomcatChartDefaultMetaInfoParams()
				Expect(releaseCache1.MetaInfoValues).To(Equal(mataInfoParams))

				By("delete release")
				err = helm.DeleteRelease(namespace, "tomcat-test")
				Expect(err).NotTo(HaveOccurred())

				err = helm.DeleteRelease(namespace, "not-existed")
				Expect(err).NotTo(HaveOccurred())

				releaseCaches, err = helm.ListAllReleases()
				Expect(err).NotTo(HaveOccurred())
				Expect(len(releaseCaches) >= 1).To(BeTrue())

				releaseCache1 = getReleaseCache(releaseCaches, namespace, "tomcat-test")
				Expect(releaseCache1).To(BeNil())
			})

		})

		Describe("test install v1 release", func() {
			var (
				zookeeperChartFiles       []*common.BufferedFile
				zookeeperV2ChartFiles     []*common.BufferedFile
				zookeeperComputedValues   map[string]interface{}
				zookeeperV2ComputedValues map[string]interface{}
			)

			BeforeEach(func() {
				zookeeperChartPath, err := framework.GetLocalV1ZookeeperChartPath()
				Expect(err).NotTo(HaveOccurred())
				zookeeperV2Chartpath, err := framework.GetLocalV2ZookeeperChartPath()
				Expect(err).NotTo(HaveOccurred())

				zookeeperChartFiles, err = framework.LoadChartArchive(zookeeperChartPath)
				Expect(err).NotTo(HaveOccurred())
				zookeeperV2ChartFiles, err = framework.LoadChartArchive(zookeeperV2Chartpath)
				Expect(err).NotTo(HaveOccurred())

				defaultValues, err := getChartDefaultValues(zookeeperChartFiles)
				Expect(err).NotTo(HaveOccurred())
				defaultV2Values, err := getChartDefaultValues(zookeeperV2ChartFiles)
				Expect(err).NotTo(HaveOccurred())

				zookeeperComputedValues = map[string]interface{}{}
				zookeeperComputedValues = util.MergeValues(zookeeperComputedValues, defaultValues, false)
				zookeeperComputedValues = util.MergeValues(zookeeperComputedValues, map[string]interface{}{
					plugins.WalmPluginConfigKey: []*k8s.ReleasePlugin{
						{
							Name: plugins.ValidateReleaseConfigPluginName,
						},
						{
							Name: plugins.IsomateSetConverterPluginName,
						},
					},
				}, false)

				zookeeperV2ComputedValues = map[string]interface{}{}
				zookeeperV2ComputedValues = util.MergeValues(zookeeperV2ComputedValues, defaultV2Values, false)
				zookeeperV2ComputedValues = util.MergeValues(zookeeperV2ComputedValues, map[string]interface{}{
					plugins.WalmPluginConfigKey: []*k8s.ReleasePlugin{
						{
							Name: plugins.ValidateReleaseConfigPluginName,
						},
						{
							Name: plugins.IsomateSetConverterPluginName,
						},
					},
				}, false)

			})

			It("test install v2 release with local v1 chart", func() {
				By("install v2 release with local v1 chart")
				releaseRequest := &release.ReleaseRequestV2{
					ReleaseRequest: release.ReleaseRequest{
						Name: "zookeeper-test",
					},
				}

				releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, zookeeperChartFiles, false, false, nil)
				Expect(err).NotTo(HaveOccurred())
				assertReleaseCacheBasic(releaseCache, namespace, "zookeeper-test", "", "zookeeper",
					"5.2.0", "5.2", 1)

				transwarpInstallID := fmt.Sprintf("%v", releaseCache.ConfigValues["Transwarp_Install_ID"])
				zookeeperComputedValues["Transwarp_Install_ID"] = transwarpInstallID
				Expect(releaseCache.ComputedValues).To(Equal(zookeeperComputedValues))

				manifest := strings.Replace("\n---\napiVersion: v1\ndata:\n  jaas.conf.tmpl: |\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    Server {\n      com.sun.security.auth.module.Krb5LoginModule required\n      useKeyTab=true\n      keyTab=\"/etc/keytabs/keytab\"\n      storeKey=true\n      useTicketCache=false\n      principal=\"{{ getv \"/security/guardian_principal_user\" \"zookeeper\" }}/{{ getv \"/security/guardian_principal_host\" \"tos\" }}@{{ getv \"/security/guardian_client_config/realm\" \"TDH\" }}\";\n    };\n    Client {\n      com.sun.security.auth.module.Krb5LoginModule required\n      useKeyTab=false\n      useTicketCache=true;\n    };\n    {{- end }}\n  log4j.properties.raw: |\n    # Define some default values that can be overridden by system properties\n    zookeeper.root.logger=INFO, CONSOLE\n    zookeeper.console.threshold=INFO\n    zookeeper.log.dir=.\n    zookeeper.log.file=zookeeper.log\n    zookeeper.log.threshold=DEBUG\n    zookeeper.tracelog.dir=.\n    zookeeper.tracelog.file=zookeeper_trace.log\n\n    #\n    # ZooKeeper Logging Configuration\n    #\n\n    # Format is \"<default threshold> (, <appender>)+\n\n    # DEFAULT: console appender only\n    log4j.rootLogger=${zookeeper.root.logger}\n\n    # Example with rolling log file\n    #log4j.rootLogger=DEBUG, CONSOLE, ROLLINGFILE\n\n    # Example with rolling log file and tracing\n    #log4j.rootLogger=TRACE, CONSOLE, ROLLINGFILE, TRACEFILE\n\n    #\n    # Log INFO level and above messages to the console\n    #\n    log4j.appender.CONSOLE=org.apache.log4j.ConsoleAppender\n    log4j.appender.CONSOLE.Threshold=${zookeeper.log.threshold}\n    log4j.appender.CONSOLE.layout=org.apache.log4j.PatternLayout\n    log4j.appender.CONSOLE.layout.ConversionPattern=%d{ISO8601} %-5p %c: [myid:%X{myid}] - [%t:%C{1}@%L] - %m%n\n\n    #\n    # Add ROLLINGFILE to rootLogger to get log file output\n    #    Log DEBUG level and above messages to a log file\n    log4j.appender.ROLLINGFILE=org.apache.log4j.RollingFileAppender\n    log4j.appender.ROLLINGFILE.Threshold=${zookeeper.log.threshold}\n    log4j.appender.ROLLINGFILE.File=${zookeeper.log.dir}/${zookeeper.log.file}\n\n    # Max log file size of 10MB\n    log4j.appender.ROLLINGFILE.MaxFileSize=64MB\n    # uncomment the next line to limit number of backup files\n    log4j.appender.ROLLINGFILE.MaxBackupIndex=4\n\n    log4j.appender.ROLLINGFILE.layout=org.apache.log4j.PatternLayout\n    log4j.appender.ROLLINGFILE.layout.ConversionPattern=%d{ISO8601} %-5p %c: [myid:%X{myid}] - [%t:%C{1}@%L] - %m%n\n\n\n    #\n    # Add TRACEFILE to rootLogger to get log file output\n    #    Log DEBUG level and above messages to a log file\n    log4j.appender.TRACEFILE=org.apache.log4j.FileAppender\n    log4j.appender.TRACEFILE.Threshold=TRACE\n    log4j.appender.TRACEFILE.File=${zookeeper.tracelog.dir}/${zookeeper.tracelog.file}\n\n    log4j.appender.TRACEFILE.layout=org.apache.log4j.PatternLayout\n    ### Notice we are including log4j's NDC here (%x)\n    log4j.appender.TRACEFILE.layout.ConversionPattern=%d{ISO8601} %-5p %c: [myid:%X{myid}] - [%t:%C{1}@%L][%x] - %m%n\n  myid.tmpl: '{{ getenv \"MYID\" }}'\n  tdh-env.sh.tmpl: |\n    #!/bin/bash\n    set -x\n\n    setup_keytab() {\n      echo \"setup_keytab\"\n    {{ if eq (getv \"/security/auth_type\") \"kerberos\" }}\n      # link_keytab\n      export KRB_MOUNTED_CONF_PATH=${KRB_MOUNTED_CONF_PATH:-/var/run/secrets/transwarp.io/tosvolume/keytab/krb5.conf}\n      export KRB_MOUNTED_KEYTAB=${KRB_MOUNTED_KEYTAB:-/var/run/secrets/transwarp.io/tosvolume/keytab/keytab}\n      if [ ! -f $KRB_MOUNTED_CONF_PATH ]; then\n        echo \"Expect krb5.conf at $KRB_MOUNTED_CONF_PATH but not found!\"\n        exit 1\n      fi\n      if [ ! -f $KRB_MOUNTED_KEYTAB ]; then\n        echo \"Expect keytab file at $KRB_MOUNTED_KEYTAB but not found!\"\n        exit 1\n      fi\n      ln -svf $KRB_MOUNTED_CONF_PATH /etc/krb5.conf\n      [ -d /etc/keytabs ] || mkdir -p /etc/keytabs\n      ln -svf $KRB_MOUNTED_KEYTAB /etc/keytabs/keytab\n    {{ end }}\n    }\n  tdh-env.toml: |-\n    [[template]]\n    src = \"tdh-env.sh.tmpl\"\n    dest = \"/etc/tdh-env.sh\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n  zoo.cfg.tmpl: |\n    # the directory where the snapshot is stored.\n    dataDir=/var/transwarp/data\n\n    # the port at which the clients will connect\n    clientPort={{ getv \"/zookeeper/zookeeper.client.port\" }}\n\n    {{- range $index, $_ := seq 0 (sub (atoi (getenv \"QUORUM_SIZE\")) 1) }}\n    server.{{ $index }}={{ getenv \"SERVICE_NAME\" }}-{{ $index }}.{{ getenv \"SERVICE_NAMESPACE\" }}.pod:{{ getv \"/zookeeper/zookeeper.peer.communicate.port\" }}:{{ getv \"/zookeeper/zookeeper.leader.elect.port\" }}\n    {{- end }}\n\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    authProvider.1=org.apache.zookeeper.server.auth.SASLAuthenticationProvider\n    jaasLoginRenew=3600000\n    kerberos.removeHostFromPrincipal=true\n    kerberos.removeRealmFromPrincipal=true\n    {{- end }}\n\n    {{- range gets \"/zoo_cfg/*\" }}\n    {{base .Key}}={{.Value}}\n    {{- end }}\n  zookeeper-confd.conf: |-\n    {\n      \"Ingress\": {\n\n      },\n      \"Transwarp_Auto_Injected_Volumes\": [\n\n      ],\n      \"msl_plugin_config\": {\n        \"config\": {\n\n        },\n        \"enable\": false\n      },\n      \"security\": {\n        \"auth_type\": \"none\",\n        \"guardian_client_config\": {\n\n        },\n        \"guardian_principal_host\": \"tos\",\n        \"guardian_principal_user\": \"zookeeper\"\n      },\n      \"zoo_cfg\": {\n        \"autopurge.purgeInterval\": 1,\n        \"autopurge.snapRetainCount\": 10,\n        \"initLimit\": 10,\n        \"maxClientCnxns\": 0,\n        \"syncLimit\": 5,\n        \"tickTime\": 9000\n      },\n      \"zookeeper\": {\n        \"zookeeper.client.port\": 2181,\n        \"zookeeper.jmxremote.port\": 9911,\n        \"zookeeper.leader.elect.port\": 3888,\n        \"zookeeper.peer.communicate.port\": 2888\n      }\n    }\n  zookeeper-env.sh.tmpl: |\n    export ZOOKEEPER_LOG_DIR=/var/transwarp/data/log\n\n    export SERVER_JVMFLAGS=\"-Dcom.sun.management.jmxremote.port={{getv \"/zookeeper/zookeeper.jmxremote.port\"}} -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.local.only=false\"\n    export SERVER_JVMFLAGS=\"-Dsun.net.inetaddr.ttl=60 -Dsun.net.inetaddr.negative.ttl=60 -Dzookeeper.refreshPeer=1 -Dzookeeper.log.dir=${ZOOKEEPER_LOG_DIR} -Dzookeeper.root.logger=INFO,CONSOLE,ROLLINGFILE $SERVER_JVMFLAGS\"\n\n    {{ if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    export SERVER_JVMFLAGS=\"-Djava.security.auth.login.config=/etc/zookeeper/conf/jaas.conf ${SERVER_JVMFLAGS}\"\n    export ZOOKEEPER_PRICIPAL={{ getv \"/security/guardian_principal_user\" \"zookeeper\" }}/{{ getv \"/security/guardian_principal_host\" \"tos\" }}@{{ getv \"/security/guardian_client_config/realm\" \"TDH\" }}\n    {{ end }}\n  zookeeper.toml: |-\n    [[template]]\n    src = \"zoo.cfg.tmpl\"\n    dest = \"/etc/zookeeper/conf/zoo.cfg\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"jaas.conf.tmpl\"\n    dest = \"/etc/zookeeper/conf/jaas.conf\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"log4j.properties.raw\"\n    dest = \"/etc/zookeeper/conf/log4j.properties\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"zookeeper-env.sh.tmpl\"\n    dest = \"/etc/zookeeper/conf/zookeeper-env.sh\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"myid.tmpl\"\n    dest = \"/var/transwarp/data/myid\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\nkind: ConfigMap\nmetadata:\n  annotations: {}\n  labels:\n    release: zookeeper-test\n    transwarp.install: 96nhl\n    transwarp.name: zookeeper-confd-conf\n  name: zookeeper-confd-conf-96nhl\n  namespace: helmreleasetest-t2295\n\n---\napiVersion: v1\ndata:\n  entrypoint.sh: |\n    #!/bin/bash\n    set -ex\n\n    export ZOOKEEPER_CONF_DIR=/etc/zookeeper/conf\n    export ZOOKEEPER_DATA_DIR=/var/transwarp\n    export ZOOKEEPER_DATA=$ZOOKEEPER_DATA_DIR/data\n    export ZOOKEEPER_CFG=$ZOOKEEPER_CONF_DIR/zoo.cfg\n\n    mkdir -p ${ZOOKEEPER_CONF_DIR}\n    mkdir -p $ZOOKEEPER_DATA\n\n    confd -onetime -backend file -prefix / -file /etc/confd/zookeeper-confd.conf\n\n    ZOOKEEPER_ENV=$ZOOKEEPER_CONF_DIR/zookeeper-env.sh\n\n    [ -f $ZOOKEEPER_ENV ] && {\n      source $ZOOKEEPER_ENV\n    }\n    [ -f /etc/tdh-env.sh ] && {\n      source /etc/tdh-env.sh\n      setup_keytab\n    }\n    # ZOOKEEPER_LOG is defined in $ZOOKEEPER_ENV\n    mkdir -p $ZOOKEEPER_LOG_DIR\n    chown -R zookeeper:zookeeper $ZOOKEEPER_LOG_DIR\n    chown -R zookeeper:zookeeper $ZOOKEEPER_DATA\n\n    echo \"Starting zookeeper service with config:\"\n    cat ${ZOOKEEPER_CFG}\n\n    sudo -u zookeeper java $SERVER_JVMFLAGS \\\n        $JAVAAGENT_OPTS \\\n        -cp $ZOOKEEPER_HOME/zookeeper-3.4.5-transwarp-with-dependencies.jar:$ZOOKEEPER_CONF_DIR \\\n        org.apache.zookeeper.server.quorum.QuorumPeerMain $ZOOKEEPER_CFG\nkind: ConfigMap\nmetadata:\n  annotations: {}\n  labels:\n    release: zookeeper-test\n    transwarp.install: 96nhl\n    transwarp.name: zookeeper-entrypoint\n  name: zookeeper-entrypoint-96nhl\n  namespace: helmreleasetest-t2295\n\n---\napiVersion: v1\nkind: Service\nmetadata:\n  annotations:\n    service.alpha.kubernetes.io/tolerate-unready-endpoints: \"true\"\n  labels:\n    k8s-app: zookeeper-hl\n    kubernetes.io/headless-service: \"true\"\n    release: zookeeper-test\n    transwarp.install: 96nhl\n    transwarp.name: zookeeper-hl\n  name: zookeeper-hl-96nhl\n  namespace: helmreleasetest-t2295\nspec:\n  clusterIP: None\n  ports:\n  - name: hl-service\n    port: 2181\n    protocol: TCP\n    targetPort: 2181\n  selector:\n    transwarp.install: 96nhl\n    transwarp.name: zookeeper\n\n---\napiVersion: v1\nkind: Service\nmetadata:\n  annotations: {}\n  labels:\n    k8s-app: zookeeper\n    kubernetes.io/cluster-service: \"true\"\n    release: zookeeper-test\n    transwarp.install: 96nhl\n    transwarp.name: zookeeper\n  name: zookeeper-96nhl\n  namespace: helmreleasetest-t2295\nspec:\n  ports:\n  - name: service\n    port: 2181\n    protocol: TCP\n    targetPort: 2181\n  selector:\n    transwarp.install: 96nhl\n    transwarp.name: zookeeper\n  type: NodePort\n\n---\napiVersion: apps/v1beta1\nkind: StatefulSet\nmetadata:\n  annotations: {}\n  labels:\n    release: zookeeper-test\n    transwarp.install: 96nhl\n    transwarp.name: zookeeper\n  name: zookeeper-96nhl\n  namespace: helmreleasetest-t2295\nspec:\n  podManagementPolicy: Parallel\n  replicas: 1\n  selector:\n    matchLabels:\n      transwarp.install: 96nhl\n      transwarp.name: zookeeper\n  serviceName: zookeeper-96nhl\n  template:\n    metadata:\n      annotations:\n        cni.networks: overlay\n        pod.alpha.kubernetes.io/initialized: \"true\"\n        transwarp/configmap.md5: 2f17b219ccca5d504959b9ccdb950fe0278a4a4894113dbc5c063780d00cab9c\n      labels:\n        release: zookeeper-test\n        transwarp.install: 96nhl\n        transwarp.name: zookeeper\n    spec:\n      affinity:\n        podAntiAffinity:\n          requiredDuringSchedulingIgnoredDuringExecution:\n          - labelSelector:\n              matchLabels:\n                release: zookeeper-test\n                transwarp.install: 96nhl\n                transwarp.name: zookeeper\n            namespaces:\n            - helmreleasetest-t2295\n            topologyKey: kubernetes.io/hostname\n      containers:\n      - args:\n        - /boot/entrypoint.sh\n        env:\n        - name: MYID\n          valueFrom:\n            fieldRef:\n              fieldPath: metadata.annotations['transwarp.replicaid']\n        - name: SERVICE_NAME\n          value: zookeeper-96nhl\n        - name: SERVICE_NAMESPACE\n          valueFrom:\n            fieldRef:\n              fieldPath: metadata.namespace\n        - name: HEAP_SIZE\n          value: 40m\n        - name: QUORUM_SIZE\n          value: \"1\"\n        image: zookeeper:transwarp-5.2\n        imagePullPolicy: Always\n        livenessProbe:\n          exec:\n            command:\n            - /bin/bash\n            - -c\n            - echo ruok|nc localhost 2181 > /dev/null && echo ok\n          initialDelaySeconds: 60\n        name: zookeeper\n        readinessProbe:\n          exec:\n            command:\n            - /bin/bash\n            - -c\n            - echo ruok|nc localhost 2181 > /dev/null && echo ok\n          initialDelaySeconds: 60\n        resources:\n          limits:\n            cpu: \"0.20000000000000001\"\n            memory: 0.040000000000000001Gi\n          requests:\n            cpu: \"0.10000000000000001\"\n            memory: 0.01Gi\n        volumeMounts:\n        - mountPath: /var/transwarp\n          name: zkdir\n        - mountPath: /boot\n          name: zookeeper-entrypoint\n        - mountPath: /etc/confd\n          name: zookeeper-confd-conf\n      hostNetwork: false\n      priorityClassName: low-priority\n      volumes:\n      - configMap:\n          items:\n          - key: entrypoint.sh\n            mode: 493\n            path: entrypoint.sh\n          name: zookeeper-entrypoint-96nhl\n        name: zookeeper-entrypoint\n      - configMap:\n          items:\n          - key: zookeeper.toml\n            path: conf.d/zookeeper.toml\n          - key: tdh-env.toml\n            path: conf.d/tdh-env.toml\n          - key: zookeeper-confd.conf\n            path: zookeeper-confd.conf\n          - key: zoo.cfg.tmpl\n            path: templates/zoo.cfg.tmpl\n          - key: jaas.conf.tmpl\n            path: templates/jaas.conf.tmpl\n          - key: zookeeper-env.sh.tmpl\n            path: templates/zookeeper-env.sh.tmpl\n          - key: myid.tmpl\n            path: templates/myid.tmpl\n          - key: log4j.properties.raw\n            path: templates/log4j.properties.raw\n          - key: tdh-env.sh.tmpl\n            path: templates/tdh-env.sh.tmpl\n          name: zookeeper-confd-conf-96nhl\n        name: zookeeper-confd-conf\n  updateStrategy:\n    type: RollingUpdate\n  volumeClaimTemplates:\n  - metadata:\n      annotations:\n        volume.beta.kubernetes.io/storage-class: silver\n      labels:\n        release: zookeeper-test\n        transwarp.install: 96nhl\n        transwarp.name: zookeeper\n      name: zkdir\n    spec:\n      accessModes:\n      - ReadWriteOnce\n      resources:\n        requests:\n          storage: 1Gi\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: zookeeper-test\n  namespace: helmreleasetest-t2295\nspec:\n  chartAppVersion: \"5.2\"\n  chartImage: \"\"\n  chartName: zookeeper\n  chartVersion: 5.2.0\n  chartWalmVersion: v1\n  configValues:\n    Transwarp_Install_ID: 96nhl\n  dependencies: null\n  dependenciesConfigValues: {}\n  isomateConfig: null\n  outputConfig:\n    provides:\n      ZOOKEEPER_CLIENT_CONFIG:\n        immediate_value:\n          zookeeper_addresses: zookeeper-96nhl-0.helmreleasetest-t2295.pod\n          zookeeper_auth_type: none\n          zookeeper_port: \"2181\"\n          zookeeper_principal: zookeeper/tos\n  repo: \"\"\nstatus: {}\n",
					"96nhl", transwarpInstallID, -1)
				manifest = strings.Replace(manifest, "helmreleasetest-t2295", namespace, -1)
				Expect(releaseCache.Manifest).To(Equal(manifest))

				Expect(releaseCache.ReleaseResourceMetas).To(Equal(getZookeeperDefaultV2ReleaseResourceMeta(namespace, "zookeeper-test", transwarpInstallID)))
				Expect(releaseCache.MetaInfoValues).To(BeNil())
			})
			//
			//// Todo: // List v1 Release. Create configmap get release from configmap
			//Describe("test list & delete v1 release", func() {
			//	var anotherNamespace string
			//
			//	BeforeEach(func() {
			//		By("create test namespace")
			//		anotherNamespace = "helmreleasetest-t2295"
			//		err = framework.CreateNamespace(anotherNamespace, nil)
			//		Expect(err).NotTo(HaveOccurred())
			//		currentFilePath, err := framework.GetCurrentFilePath()
			//		Expect(err).NotTo(HaveOccurred())
			//
			//		appmanagerConfigMapPath := filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/v1/zookeeper/appmanager-configmap.yaml")
			//		confdConfigMapPath := filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/v1/zookeeper/zookeeper-confd-conf.yaml")
			//		entrypointConfigMapPath := filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/v1/zookeeper/zookeeper-entrypoint.yaml")
			//		releaseConfigMapPath := filepath.Join(filepath.Dir(currentFilePath), "../../resources/helm/v1/zookeeper/configmap.yaml")
			//
			//		_, err = framework.CreateCustomConfigMap(anotherNamespace, appmanagerConfigMapPath)
			//		Expect(err).NotTo(HaveOccurred())
			//
			//		_, err = framework.CreateCustomConfigMap(anotherNamespace, confdConfigMapPath)
			//		Expect(err).NotTo(HaveOccurred())
			//
			//		_, err = framework.CreateCustomConfigMap(anotherNamespace, entrypointConfigMapPath)
			//		Expect(err).NotTo(HaveOccurred())
			//
			//		_, err = framework.CreateCustomConfigMap(anotherNamespace, releaseConfigMapPath)
			//		Expect(err).NotTo(HaveOccurred())
			//	})
			//
			//	AfterEach(func() {
			//		By("delete test namespace")
			//		err = framework.DeleteNamespace(anotherNamespace)
			//		Expect(err).NotTo(HaveOccurred())
			//	})
			//	It("list & delete v1 release", func() {
			//		if setting.Config.CrdConfig.NotNeedInstance {
			//			By("Ignore testing list & delete v1 release due to disabled ApplicationInstance")
			//			return
			//		}
			//		By("list v1 release")
			//		releaseCaches, err := helm.ListAllReleases()
			//		Expect(err).NotTo(HaveOccurred())
			//		getReleaseCache := func(releaseCaches []*release.ReleaseCache, anoterNamespace, name string) *release.ReleaseCache {
			//			for _, releaseCache := range releaseCaches {
			//				if releaseCache.Name == "helmreleasetest-zk" && releaseCache.Namespace == anoterNamespace {
			//					return releaseCache
			//				}
			//			}
			//			return nil
			//		}
			//		releaseCache := getReleaseCache(releaseCaches, anotherNamespace, "helmreleasetest-zk")
			//		Expect(releaseCache).NotTo(BeNil())
			//		assertReleaseCacheBasic(releaseCache, anotherNamespace, "helmreleasetest-zk", "", "zookeeper",
			//			"5.2.0", "5.2", 1)
			//
			//		Expect(releaseCache.HelmVersion).To(Equal("v2"))
			//		Expect(releaseCache.ReleaseResourceMetas).To(Equal(getZookeeperDefaultV1ReleaseResourceMeta(anotherNamespace, "helmreleasetest-zk")))
			//
			//		By("delete v1 release")
			//		err = helm.DeleteRelease(namespace, "helmreleasetest-zk")
			//		Expect(err).NotTo(HaveOccurred())
			//		err = helm.DeleteRelease(namespace, "not-existed")
			//		Expect(err).NotTo(HaveOccurred())
			//
			//		releaseCaches, err = helm.ListAllReleases()
			//		Expect(err).NotTo(HaveOccurred())
			//		releaseCache = getReleaseCache(releaseCaches, namespace, "helmreleasetest-zk")
			//		Expect(releaseCache).To(BeNil())
			//	})
			//})
		})

	})

})

func assertYamlConfigValues(expectedValues map[string]interface{}, Values map[string]interface{}) {
	serializedExpectedValues, err := serializeYamlConfigValues(expectedValues)
	Expect(err).NotTo(HaveOccurred())

	serializedValues, err := serializeYamlConfigValues(Values)
	Expect(err).NotTo(HaveOccurred())

	Expect(serializedExpectedValues).To(Equal(serializedValues))
}

func serializeYamlConfigValues(values map[string]interface{}) (map[string]interface{}, error) {
	valueStr, err := yaml.Marshal(values)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{}
	err = yaml.Unmarshal(valueStr, &result)
	return result, err
}

func getTomcatChartDefaultReleaseResourceMeta(namespace, name string) []release.ReleaseResourceMeta {
	return []release.ReleaseResourceMeta{
		{
			Namespace: namespace,
			Name:      name,
			Kind:      "Service",
		},
		{
			Namespace: namespace,
			Name:      name,
			Kind:      "Deployment",
		},
		{
			Namespace: namespace,
			Name:      name,
			Kind:      "ReleaseConfig",
		},
	}
}

func getZookeeperDefaultReleaseResourceMeta(namespace, name string) []release.ReleaseResourceMeta {
	return []release.ReleaseResourceMeta{
		{
			Kind:      "ConfigMap",
			Namespace: namespace,
			Name:      name + "-zookeeper-confd-conf",
		},
		{
			Kind:      "ConfigMap",
			Namespace: namespace,
			Name:      name + "-zookeeper-entrypoint",
		},
		{
			Kind:      "Service",
			Namespace: namespace,
			Name:      name + "-zookeeper-hl",
		},
		{
			Kind:      "StatefulSet",
			Namespace: namespace,
			Name:      name + "-zookeeper",
		},
		{
			Kind:      "ReleaseConfig",
			Namespace: namespace,
			Name:      name,
		},
	}
}

func getZookeeperDefaultV1ReleaseResourceMeta(namespace, name string) []release.ReleaseResourceMeta {
	return []release.ReleaseResourceMeta{
		{
			Kind:      "ConfigMap",
			Namespace: namespace,
			Name:      "appmanager." + name + ".1778257097",
		},
		{
			Kind:      "ApplicationInstance",
			Namespace: namespace,
			Name:      name,
		},
	}
}

func getZookeeperDefaultV2ReleaseResourceMeta(namespace, name, transwarpInstallID string) []release.ReleaseResourceMeta {
	return []release.ReleaseResourceMeta{
		{
			Kind:      "ConfigMap",
			Namespace: namespace,
			Name:      "zookeeper-confd-conf-" + transwarpInstallID,
		},
		{
			Kind:      "ConfigMap",
			Namespace: namespace,
			Name:      "zookeeper-entrypoint-" + transwarpInstallID,
		},
		{
			Kind:      "Service",
			Namespace: namespace,
			Name:      "zookeeper-hl-" + transwarpInstallID,
		},
		{
			Kind:      "Service",
			Namespace: namespace,
			Name:      "zookeeper-" + transwarpInstallID,
		},
		{
			Kind:      "StatefulSet",
			Namespace: namespace,
			Name:      "zookeeper-" + transwarpInstallID,
		},
		{
			Kind:      "ReleaseConfig",
			Namespace: namespace,
			Name:      name,
		},
	}
}

func assertReleaseCacheBasic(cache *release.ReleaseCache, namespace, name, repo, chartName, chartVersion,
	chartAppVersion string, version int32) {

	Expect(cache.Name).To(Equal(name))
	Expect(cache.Namespace).To(Equal(namespace))
	Expect(cache.RepoName).To(Equal(repo))
	Expect(cache.ChartName).To(Equal(chartName))
	Expect(cache.ChartVersion).To(Equal(chartVersion))
	Expect(cache.ChartAppVersion).To(Equal(chartAppVersion))
	Expect(cache.Version).To(Equal(version))
}

func getChartDefaultValues(chartFiles []*common.BufferedFile) (map[string]interface{}, error) {
	for _, file := range chartFiles {
		if file.Name == "values.yaml" {
			defaultValues := map[string]interface{}{}
			err := yaml.Unmarshal(file.Data, &defaultValues)
			return defaultValues, err
		}
	}
	return nil, errors.New("values.yaml is not found")
}

func getTomcatChartDefaultMetaInfoParams() *release.MetaInfoParams {
	webarchiveImage := "ananwaresystems/webarchive:1.0"
	tomcatImage := "tomcat:7.0"
	tomcatReplicas := int64(1)
	tomcatLimitsMemory := int64(200)
	tomcatLimitsCpu := float64(0.2)
	tomcatRequestsMemory := int64(100)
	tomcatRequestsCpu := float64(0.1)
	return &release.MetaInfoParams{
		Roles: []*release.MetaRoleConfigValue{
			{
				Name: "webarchive",
				RoleBaseConfigValue: &release.MetaRoleBaseConfigValue{
					Image: &webarchiveImage,
				},
			},
			{
				Name: "tomcat",
				RoleBaseConfigValue: &release.MetaRoleBaseConfigValue{
					Image:    &tomcatImage,
					Replicas: &tomcatReplicas,
					Others: []*release.MetaCommonConfigValue{
						{
							Name:  "path",
							Type:  "string",
							Value: "\"/sample\"",
						},
					},
				},
				RoleResourceConfigValue: &release.MetaResourceConfigValue{
					LimitsMemory:   &tomcatLimitsMemory,
					LimitsCpu:      &tomcatLimitsCpu,
					RequestsMemory: &tomcatRequestsMemory,
					RequestsCpu:    &tomcatRequestsCpu,
				},
			},
		},
	}
}

func getZookeeperDefaultMetaInfoParams() *release.MetaInfoParams {
	tomcatImage := "docker.io/corndai1997/zookeeper:5.2"
	tomcatReplicas := int64(3)
	priority := int64(0)
	useHostNetwork := false
	envs := []release.MetaEnv{}
	tomcatLimitsMemory := int64(200)
	tomcatLimitsCpu := float64(0.2)
	tomcatRequestsMemory := int64(100)
	tomcatRequestsCpu := float64(0.1)
	return &release.MetaInfoParams{
		Roles: []*release.MetaRoleConfigValue{
			{
				Name: "zookeeper",
				RoleBaseConfigValue: &release.MetaRoleBaseConfigValue{
					Image:          &tomcatImage,
					Replicas:       &tomcatReplicas,
					Priority:       &priority,
					UseHostNetwork: &useHostNetwork,
					Env:            envs,
				},
				RoleResourceConfigValue: &release.MetaResourceConfigValue{
					LimitsMemory:   &tomcatLimitsMemory,
					LimitsCpu:      &tomcatLimitsCpu,
					RequestsMemory: &tomcatRequestsMemory,
					RequestsCpu:    &tomcatRequestsCpu,
					StorageResources: []*release.MetaResourceStorageConfigValue{
						{
							Name: "data",
							Value: &release.MetaResourceStorage{
								ResourceStorage: release.ResourceStorage{
									StorageClass: "local",
								},
								Size: 100,
							},
						},
					},
				},
			},
		},
		Params: []*release.MetaCommonConfigValue{
			{
				Name:  "zoo_cfg",
				Type:  "kvPair",
				Value: "{\"autopurge.purgeInterval\":5,\"autopurge.snapRetainCount\":10,\"initLimit\":10,\"maxClientCnxns\":0,\"syncLimit\":5,\"tickTime\":9000}",
			},
			{
				Name:  "zookeeper",
				Type:  "kvPair",
				Value: "{\"zookeeper.client.port\":2181,\"zookeeper.jmxremote.port\":9911,\"zookeeper.leader.elect.port\":3888,\"zookeeper.peer.communicate.port\":2888}",
			},
		},
	}
}
