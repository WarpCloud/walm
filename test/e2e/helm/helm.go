package helm

import (
	. "github.com/onsi/gomega"
	. "github.com/onsi/ginkgo"
	"WarpCloud/walm/test/e2e/framework"
	"WarpCloud/walm/pkg/k8s/cache/informer"
	"WarpCloud/walm/pkg/helm/impl"
	"WarpCloud/walm/pkg/setting"
	"WarpCloud/walm/pkg/models/release"
	"WarpCloud/walm/pkg/models/common"
	"github.com/ghodss/yaml"
	"errors"
	"WarpCloud/walm/pkg/util"
	"k8s.io/helm/pkg/walm"
	"k8s.io/helm/pkg/walm/plugins"
	"fmt"
	"path/filepath"
	"strings"
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
		k8sCache := informer.NewInformer(framework.GetK8sClient(), framework.GetK8sReleaseConfigClient(), 0, stopChan)
		registryClient := impl.NewRegistryClient(setting.Config.ChartImageConfig)

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
			tomcatChartPath, err := framework.GetTestTomcatChartPath()
			Expect(err).NotTo(HaveOccurred())

			tomcatChartFiles, err = framework.LoadChartArchive(tomcatChartPath)
			Expect(err).NotTo(HaveOccurred())

			defaultValues, err := getChartDefaultValues(tomcatChartFiles)
			Expect(err).NotTo(HaveOccurred())
			tomcatComputedValues = map[string]interface{}{}
			tomcatComputedValues = util.MergeValues(tomcatComputedValues, defaultValues, false)
			tomcatComputedValues = util.MergeValues(tomcatComputedValues, map[string]interface{}{
				walm.WalmPluginConfigKey: []*walm.WalmPlugin{
					{
						Name: plugins.ValidateReleaseConfigPluginName,
					},
				},
			}, false)
		})

		It("test local chart", func() {
			By("install release with local chart")
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "tomcat-test",
				},
			}

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, tomcatChartFiles, false, false, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 1)

			Expect(releaseCache.ComputedValues).To(Equal(tomcatComputedValues))

			manifest := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartImage: \"\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  configValues: null\n  dependencies: null\n  dependenciesConfigValues: {}\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n",
				namespace, namespace, namespace)
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

			mataInfoParams := getTomcatChartDefaultMetaInfoParams()
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})

		It("test repo chart", func() {
			By("install release with repo chart")
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name:         "tomcat-test",
					RepoName:     framework.TestChartRepoName,
					ChartName:    framework.TestChartName,
					ChartVersion: framework.TestChartVersion,
				},
			}

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, nil, false, false, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 1)

			Expect(releaseCache.ComputedValues).To(Equal(tomcatComputedValues))

			manifest := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartImage: \"\"\n  chartName: tomcat\n  chartVersion: 0.2.0\n  configValues: null\n  dependencies: null\n  dependenciesConfigValues: {}\n  outputConfig: {}\n  repo: %s\nstatus: {}\n",
				namespace, namespace, namespace, framework.TestChartRepoName)
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

			mataInfoParams := getTomcatChartDefaultMetaInfoParams()
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})

		It("test chart image", func() {
			By("install release with chart image")
			releaseRequest := &release.ReleaseRequestV2{
				ReleaseRequest: release.ReleaseRequest{
					Name: "tomcat-test",
				},
				ChartImage: framework.GetTestChartImage(),
			}

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, nil, false, false, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			assertReleaseCacheBasic(releaseCache, namespace, "tomcat-test", "", "tomcat",
				"0.2.0", "7", 1)

			Expect(releaseCache.ComputedValues).To(Equal(tomcatComputedValues))

			manifest := fmt.Sprintf("\n---\napiVersion: v1\nkind: Service\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  ports:\n  - name: http\n    port: 80\n    protocol: TCP\n    targetPort: 8080\n  selector:\n    app: tomcat\n    release: tomcat-test\n  type: NodePort\n\n---\napiVersion: apps/v1beta2\nkind: Deployment\nmetadata:\n  labels:\n    app: tomcat\n    chart: tomcat-0.2.0\n    heritage: Helm\n    release: tomcat-test\n  name: tomcat-test\n  namespace: %s\nspec:\n  replicas: 1\n  selector:\n    matchLabels:\n      app: tomcat\n      release: tomcat-test\n  template:\n    metadata:\n      labels:\n        app: tomcat\n        release: tomcat-test\n    spec:\n      containers:\n      - image: tomcat:7.0\n        imagePullPolicy: Always\n        livenessProbe:\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        name: tomcat\n        ports:\n        - containerPort: 8080\n          hostPort: 8009\n        readinessProbe:\n          failureThreshold: 6\n          httpGet:\n            path: /sample\n            port: 8080\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: 0.2\n            memory: 200Mi\n          requests:\n            cpu: 0.1\n            memory: 100Mi\n        volumeMounts:\n        - mountPath: /usr/local/tomcat/webapps\n          name: app-volume\n      initContainers:\n      - command:\n        - sh\n        - -c\n        - cp /*.war /app\n        image: ananwaresystems/webarchive:1.0\n        imagePullPolicy: Always\n        name: war\n        volumeMounts:\n        - mountPath: /app\n          name: app-volume\n      volumes:\n      - emptyDir: {}\n        name: app-volume\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  name: tomcat-test\n  namespace: %s\nspec:\n  chartAppVersion: \"7\"\n  chartImage: %s\n  chartName: tomcat\n  chartVersion: 0.2.0\n  configValues: null\n  dependencies: null\n  dependenciesConfigValues: {}\n  outputConfig: {}\n  repo: \"\"\nstatus: {}\n",
				namespace, namespace, namespace, framework.GetTestChartImage())
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getTomcatChartDefaultReleaseResourceMeta(namespace, "tomcat-test")))

			mataInfoParams := getTomcatChartDefaultMetaInfoParams()
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})

		It("test jsonnet chart", func() {
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

			releaseCache, err := helm.InstallOrCreateRelease(namespace, releaseRequest, chartFiles, false, false, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			assertReleaseCacheBasic(releaseCache, namespace, "zookeeper-test", "", "zookeeper",
				"6.1.0", "6.1", 1)

			defaultValues, err := getChartDefaultValues(chartFiles)
			Expect(err).NotTo(HaveOccurred())
			computedValues := map[string]interface{}{}
			computedValues = util.MergeValues(computedValues, defaultValues, false)
			computedValues = util.MergeValues(computedValues, map[string]interface{}{
				walm.WalmPluginConfigKey: []*walm.WalmPlugin{
					{
						Name: plugins.ValidateReleaseConfigPluginName,
					},
				},
			}, false)

			Expect(releaseCache.ComputedValues).To(Equal(computedValues))

			manifest := strings.Replace("\n---\napiVersion: v1\ndata:\n  jaas.conf.tmpl: |\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    Server {\n      com.sun.security.auth.module.Krb5LoginModule required\n      useKeyTab=true\n      keyTab=\"/etc/keytabs/keytab\"\n      storeKey=true\n      useTicketCache=false\n      principal=\"{{ getv \"/security/guardian_principal_user\" \"zookeeper\" }}/{{ getv \"/security/guardian_principal_host\" \"tos\" }}@{{ getv \"/security/guardian_client_config/realm\" \"TDH\" }}\";\n    };\n    Client {\n      com.sun.security.auth.module.Krb5LoginModule required\n      useKeyTab=false\n      useTicketCache=true;\n    };\n    {{- end }}\n  log4j.properties.raw: |\n    # Define some default values that can be overridden by system properties\n    zookeeper.root.logger=INFO, CONSOLE\n    zookeeper.console.threshold=INFO\n    zookeeper.log.dir=.\n    zookeeper.log.file=zookeeper.log\n    zookeeper.log.threshold=DEBUG\n    zookeeper.tracelog.dir=.\n    zookeeper.tracelog.file=zookeeper_trace.log\n\n    #\n    # ZooKeeper Logging Configuration\n    #\n\n    # Format is \"<default threshold> (, <appender>)+\n\n    # DEFAULT: console appender only\n    log4j.rootLogger=${zookeeper.root.logger}\n\n    # Example with rolling log file\n    #log4j.rootLogger=DEBUG, CONSOLE, ROLLINGFILE\n\n    # Example with rolling log file and tracing\n    #log4j.rootLogger=TRACE, CONSOLE, ROLLINGFILE, TRACEFILE\n\n    #\n    # Log INFO level and above messages to the console\n    #\n    log4j.appender.CONSOLE=org.apache.log4j.ConsoleAppender\n    log4j.appender.CONSOLE.Threshold=${zookeeper.log.threshold}\n    log4j.appender.CONSOLE.layout=org.apache.log4j.PatternLayout\n    log4j.appender.CONSOLE.layout.ConversionPattern=%d{ISO8601} %-5p %c: [myid:%X{myid}] - [%t:%C{1}@%L] - %m%n\n\n    #\n    # Add ROLLINGFILE to rootLogger to get log file output\n    #    Log DEBUG level and above messages to a log file\n    log4j.appender.ROLLINGFILE=org.apache.log4j.RollingFileAppender\n    log4j.appender.ROLLINGFILE.Threshold=${zookeeper.log.threshold}\n    log4j.appender.ROLLINGFILE.File=${zookeeper.log.dir}/${zookeeper.log.file}\n\n    # Max log file size of 10MB\n    log4j.appender.ROLLINGFILE.MaxFileSize=64MB\n    # uncomment the next line to limit number of backup files\n    log4j.appender.ROLLINGFILE.MaxBackupIndex=4\n\n    log4j.appender.ROLLINGFILE.layout=org.apache.log4j.PatternLayout\n    log4j.appender.ROLLINGFILE.layout.ConversionPattern=%d{ISO8601} %-5p %c: [myid:%X{myid}] - [%t:%C{1}@%L] - %m%n\n\n\n    #\n    # Add TRACEFILE to rootLogger to get log file output\n    #    Log DEBUG level and above messages to a log file\n    log4j.appender.TRACEFILE=org.apache.log4j.FileAppender\n    log4j.appender.TRACEFILE.Threshold=TRACE\n    log4j.appender.TRACEFILE.File=${zookeeper.tracelog.dir}/${zookeeper.tracelog.file}\n\n    log4j.appender.TRACEFILE.layout=org.apache.log4j.PatternLayout\n    ### Notice we are including log4j's NDC here (%x)\n    log4j.appender.TRACEFILE.layout.ConversionPattern=%d{ISO8601} %-5p %c: [myid:%X{myid}] - [%t:%C{1}@%L][%x] - %m%n\n  myid.tmpl: '{{ getenv \"MYID\" }}'\n  tdh-env.sh.tmpl: |\n    #!/bin/bash\n    set -x\n\n    setup_keytab() {\n      echo \"setup_keytab\"\n    {{ if eq (getv \"/security/auth_type\") \"kerberos\" }}\n      # link_keytab\n      export KRB_MOUNTED_CONF_PATH=${KRB_MOUNTED_CONF_PATH:-/var/run/secrets/transwarp.io/tosvolume/keytab/krb5.conf}\n      export KRB_MOUNTED_KEYTAB=${KRB_MOUNTED_KEYTAB:-/var/run/secrets/transwarp.io/tosvolume/keytab/keytab}\n      if [ ! -f $KRB_MOUNTED_CONF_PATH ]; then\n        echo \"Expect krb5.conf at $KRB_MOUNTED_CONF_PATH but not found!\"\n        exit 1\n      fi\n      if [ ! -f $KRB_MOUNTED_KEYTAB ]; then\n        echo \"Expect keytab file at $KRB_MOUNTED_KEYTAB but not found!\"\n        exit 1\n      fi\n      ln -svf $KRB_MOUNTED_CONF_PATH /etc/krb5.conf\n      [ -d /etc/keytabs ] || mkdir -p /etc/keytabs\n      ln -svf $KRB_MOUNTED_KEYTAB /etc/keytabs/keytab\n    {{ end }}\n    }\n  tdh-env.toml: |-\n    [[template]]\n    src = \"tdh-env.sh.tmpl\"\n    dest = \"/etc/tdh-env.sh\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n  zoo.cfg.tmpl: |\n    # the directory where the snapshot is stored.\n    dataDir=/var/transwarp/data\n\n    # the port at which the clients will connect\n    clientPort={{ getv \"/zookeeper/zookeeper.client.port\" }}\n\n    {{- range $index, $_ := seq 0 (sub (atoi (getenv \"QUORUM_SIZE\")) 1) }}\n    server.{{ $index }}={{ getenv \"SERVICE_NAME\" }}-{{ $index }}.{{ getenv \"SERVICE_NAME\" }}-hl.{{ getenv \"SERVICE_NAMESPACE\" }}.svc:{{ getv \"/zookeeper/zookeeper.peer.communicate.port\" }}:{{ getv \"/zookeeper/zookeeper.leader.elect.port\" }}\n    {{- end }}\n\n    {{- if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    authProvider.1=org.apache.zookeeper.server.auth.SASLAuthenticationProvider\n    jaasLoginRenew=3600000\n    kerberos.removeHostFromPrincipal=true\n    kerberos.removeRealmFromPrincipal=true\n    {{- end }}\n\n    {{- range gets \"/zoo_cfg/*\" }}\n    {{base .Key}}={{.Value}}\n    {{- end }}\n  zookeeper-confd.conf: |-\n    {\n      \"security\": {\n        \"auth_type\": \"none\"\n      },\n      \"transwarpApplicationPause\": false,\n      \"transwarpCniNetwork\": \"overlay\",\n      \"transwarpGlobalIngress\": {\n        \"httpPort\": 80,\n        \"httpsPort\": 443\n      },\n      \"transwarpLicenseAddress\": \"\",\n      \"transwarpMetrics\": {\n        \"enable\": true\n      },\n      \"zoo_cfg\": {\n        \"autopurge.purgeInterval\": 5,\n        \"autopurge.snapRetainCount\": 10,\n        \"initLimit\": 10,\n        \"maxClientCnxns\": 0,\n        \"syncLimit\": 5,\n        \"tickTime\": 9000\n      },\n      \"zookeeper\": {\n        \"zookeeper.client.port\": 2181,\n        \"zookeeper.jmxremote.port\": 9911,\n        \"zookeeper.leader.elect.port\": 3888,\n        \"zookeeper.peer.communicate.port\": 2888\n      }\n    }\n  zookeeper-env.sh.tmpl: |\n    export ZOOKEEPER_LOG_DIR=/var/transwarp/data/log\n\n    export SERVER_JVMFLAGS=\"-Dcom.sun.management.jmxremote.port={{getv \"/zookeeper/zookeeper.jmxremote.port\"}} -Dcom.sun.management.jmxremote.authenticate=false -Dcom.sun.management.jmxremote.ssl=false -Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.local.only=false\"\n    export SERVER_JVMFLAGS=\"-Dsun.net.inetaddr.ttl=60 -Dsun.net.inetaddr.negative.ttl=60 -Dzookeeper.refreshPeer=1 -Dzookeeper.log.dir=${ZOOKEEPER_LOG_DIR} -Dzookeeper.root.logger=INFO,CONSOLE,ROLLINGFILE $SERVER_JVMFLAGS\"\n\n    {{ if eq (getv \"/security/auth_type\") \"kerberos\" }}\n    export SERVER_JVMFLAGS=\"-Djava.security.auth.login.config=/etc/zookeeper/conf/jaas.conf ${SERVER_JVMFLAGS}\"\n    export ZOOKEEPER_PRICIPAL={{ getv \"/security/guardian_principal_user\" \"zookeeper\" }}/{{ getv \"/security/guardian_principal_host\" \"tos\" }}@{{ getv \"/security/guardian_client_config/realm\" \"TDH\" }}\n    {{ end }}\n  zookeeper.toml: |-\n    [[template]]\n    src = \"zoo.cfg.tmpl\"\n    dest = \"/etc/zookeeper/conf/zoo.cfg\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"jaas.conf.tmpl\"\n    dest = \"/etc/zookeeper/conf/jaas.conf\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"log4j.properties.raw\"\n    dest = \"/etc/zookeeper/conf/log4j.properties\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"zookeeper-env.sh.tmpl\"\n    dest = \"/etc/zookeeper/conf/zookeeper-env.sh\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\n\n    [[template]]\n    src = \"myid.tmpl\"\n    dest = \"/var/transwarp/data/myid\"\n    check_cmd = \"/bin/true\"\n    reload_cmd = \"/bin/true\"\n    keys = [ \"/\" ]\nkind: ConfigMap\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: zookeeper\n    app.kubernetes.io/instance: zookeeper-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: zookeeper-confd-conf\n    app.kubernetes.io/part-of: zookeeper\n    app.kubernetes.io/version: \"6.1\"\n  name: zookeeper-test-zookeeper-confd-conf\n  namespace: helmreleasetest-67t9g\n\n---\napiVersion: v1\ndata:\n  entrypoint.sh: |\n    #!/bin/bash\n    set -ex\n\n    export ZOOKEEPER_CONF_DIR=/etc/zookeeper/conf\n    export ZOOKEEPER_DATA_DIR=/var/transwarp\n    export ZOOKEEPER_DATA=$ZOOKEEPER_DATA_DIR/data\n    export ZOOKEEPER_CFG=$ZOOKEEPER_CONF_DIR/zoo.cfg\n\n    mkdir -p ${ZOOKEEPER_CONF_DIR}\n    mkdir -p $ZOOKEEPER_DATA\n\n    export MYID=${HOSTNAME##*-}\n    confd -onetime -backend file -prefix / -file /etc/confd/zookeeper-confd.conf\n\n    ZOOKEEPER_ENV=$ZOOKEEPER_CONF_DIR/zookeeper-env.sh\n\n    [ -f $ZOOKEEPER_ENV ] && {\n      source $ZOOKEEPER_ENV\n    }\n    [ -f /etc/tdh-env.sh ] && {\n      source /etc/tdh-env.sh\n      setup_keytab\n    }\n    # ZOOKEEPER_LOG is defined in $ZOOKEEPER_ENV\n    mkdir -p $ZOOKEEPER_LOG_DIR\n    chown -R zookeeper:zookeeper $ZOOKEEPER_LOG_DIR\n    chown -R zookeeper:zookeeper $ZOOKEEPER_DATA\n\n    echo \"Starting zookeeper service with config:\"\n    cat ${ZOOKEEPER_CFG}\n\n    JMXEXPORTER_ENABLED=${JMXEXPORTER_ENABLED:-\"true\"}\n    if [ \"${JMXEXPORTER_ENABLED}\" == \"true\" ];then\n      export JAVAAGENT_OPTS=\" -javaagent:/usr/lib/jmx_exporter/jmx_prometheus_javaagent-0.7.jar=19000:/usr/lib/jmx_exporter/agentconfig.yml \"\n    fi\n\n    sudo -u zookeeper java $SERVER_JVMFLAGS \\\n        $JAVAAGENT_OPTS \\\n        -cp $ZOOKEEPER_HOME/zookeeper-3.4.5-transwarp-with-dependencies.jar:$ZOOKEEPER_CONF_DIR \\\n        org.apache.zookeeper.server.quorum.QuorumPeerMain $ZOOKEEPER_CFG\nkind: ConfigMap\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: zookeeper\n    app.kubernetes.io/instance: zookeeper-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: zookeeper-entrypoint\n    app.kubernetes.io/part-of: zookeeper\n    app.kubernetes.io/version: \"6.1\"\n  name: zookeeper-test-zookeeper-entrypoint\n  namespace: helmreleasetest-67t9g\n\n---\napiVersion: v1\nkind: Service\nmetadata:\n  annotations:\n    service.alpha.kubernetes.io/tolerate-unready-endpoints: \"true\"\n  labels:\n    app.kubernetes.io/component: zookeeper\n    app.kubernetes.io/instance: zookeeper-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: zookeeper\n    app.kubernetes.io/part-of: zookeeper\n    app.kubernetes.io/service-type: headless-service\n    app.kubernetes.io/version: \"6.1\"\n  name: zookeeper-test-zookeeper-hl\n  namespace: helmreleasetest-67t9g\nspec:\n  clusterIP: None\n  ports:\n  - name: zk-port\n    port: 2181\n    targetPort: 2181\n  selector:\n    app.kubernetes.io/instance: zookeeper-test\n    app.kubernetes.io/name: zookeeper\n    app.kubernetes.io/version: \"6.1\"\n\n---\napiVersion: apps/v1beta1\nkind: StatefulSet\nmetadata:\n  annotations: {}\n  labels:\n    app.kubernetes.io/component: zookeeper\n    app.kubernetes.io/instance: zookeeper-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: zookeeper\n    app.kubernetes.io/part-of: zookeeper\n    app.kubernetes.io/version: \"6.1\"\n  name: zookeeper-test-zookeeper\n  namespace: helmreleasetest-67t9g\nspec:\n  podManagementPolicy: Parallel\n  replicas: 3\n  selector:\n    matchLabels:\n      app.kubernetes.io/instance: zookeeper-test\n      app.kubernetes.io/name: zookeeper\n      app.kubernetes.io/version: \"6.1\"\n  serviceName: zookeeper-test-zookeeper-hl\n  template:\n    metadata:\n      annotations:\n        tos.network.staticIP: \"true\"\n      labels:\n        app.kubernetes.io/instance: zookeeper-test\n        app.kubernetes.io/name: zookeeper\n        app.kubernetes.io/version: \"6.1\"\n    spec:\n      affinity:\n        podAntiAffinity:\n          requiredDuringSchedulingIgnoredDuringExecution:\n          - labelSelector:\n              matchLabels:\n                app.kubernetes.io/instance: zookeeper-test\n                app.kubernetes.io/name: zookeeper\n                app.kubernetes.io/version: \"6.1\"\n            namespaces:\n            - helmreleasetest-67t9g\n            topologyKey: kubernetes.io/hostname\n      containers:\n      - command:\n        - /boot/entrypoint.sh\n        env:\n        - name: SERVICE_NAME\n          value: zookeeper-test-zookeeper\n        - name: SERVICE_NAMESPACE\n          value: helmreleasetest-67t9g\n        - name: QUORUM_SIZE\n          value: \"3\"\n        image: docker.io/corndai1997/zookeeper:5.2\n        imagePullPolicy: Always\n        name: zookeeper\n        readinessProbe:\n          exec:\n            command:\n            - /bin/bash\n            - -c\n            - echo ruok|nc localhost 2181 > /dev/null && echo ok\n          initialDelaySeconds: 60\n          periodSeconds: 30\n        resources:\n          limits:\n            cpu: \"0.20000000000000001\"\n            memory: 200Mi\n            nvidia.com/gpu: \"0\"\n          requests:\n            cpu: \"0.10000000000000001\"\n            memory: 100Mi\n            nvidia.com/gpu: \"0\"\n        volumeMounts:\n        - mountPath: /boot\n          name: zookeeper-entrypoint\n        - mountPath: /etc/confd\n          name: zookeeper-confd-conf\n        - mountPath: /var/transwarp\n          name: zkdir\n      hostNetwork: false\n      initContainers: []\n      priorityClassName: low-priority\n      restartPolicy: Always\n      terminationGracePeriodSeconds: 30\n      volumes:\n      - configMap:\n          items:\n          - key: entrypoint.sh\n            mode: 493\n            path: entrypoint.sh\n          name: zookeeper-test-zookeeper-entrypoint\n        name: zookeeper-entrypoint\n      - configMap:\n          items:\n          - key: zookeeper.toml\n            path: conf.d/zookeeper.toml\n          - key: tdh-env.toml\n            path: conf.d/tdh-env.toml\n          - key: zookeeper-confd.conf\n            path: zookeeper-confd.conf\n          - key: zoo.cfg.tmpl\n            path: templates/zoo.cfg.tmpl\n          - key: jaas.conf.tmpl\n            path: templates/jaas.conf.tmpl\n          - key: zookeeper-env.sh.tmpl\n            path: templates/zookeeper-env.sh.tmpl\n          - key: myid.tmpl\n            path: templates/myid.tmpl\n          - key: log4j.properties.raw\n            path: templates/log4j.properties.raw\n          - key: tdh-env.sh.tmpl\n            path: templates/tdh-env.sh.tmpl\n          name: zookeeper-test-zookeeper-confd-conf\n        name: zookeeper-confd-conf\n  updateStrategy:\n    type: RollingUpdate\n  volumeClaimTemplates:\n  - metadata:\n      annotations:\n        volume.beta.kubernetes.io/storage-class: local\n      labels:\n        app.kubernetes.io/component: zookeeper\n        app.kubernetes.io/instance: zookeeper-test\n        app.kubernetes.io/managed-by: walm\n        app.kubernetes.io/name: zookeeper\n        app.kubernetes.io/part-of: zookeeper\n        app.kubernetes.io/version: \"6.1\"\n      name: zkdir\n    spec:\n      accessModes:\n      - ReadWriteOnce\n      resources:\n        requests:\n          storage: 100Gi\n      storageClassName: local\n\n---\napiVersion: apiextensions.transwarp.io/v1beta1\nkind: ReleaseConfig\nmetadata:\n  creationTimestamp: null\n  labels:\n    app.kubernetes.io/component: zookeeper\n    app.kubernetes.io/instance: zookeeper-test\n    app.kubernetes.io/managed-by: walm\n    app.kubernetes.io/name: zookeeper\n    app.kubernetes.io/part-of: zookeeper\n    app.kubernetes.io/version: \"6.1\"\n  name: zookeeper-test\n  namespace: helmreleasetest-67t9g\nspec:\n  chartAppVersion: \"6.1\"\n  chartImage: \"\"\n  chartName: zookeeper\n  chartVersion: 6.1.0\n  configValues: null\n  dependencies: null\n  dependenciesConfigValues: {}\n  outputConfig:\n    zookeeper_addresses: zookeeper-test-zookeeper-0.zookeeper-test-zookeeper-hl.helmreleasetest-67t9g.svc,zookeeper-test-zookeeper-1.zookeeper-test-zookeeper-hl.helmreleasetest-67t9g.svc,zookeeper-test-zookeeper-2.zookeeper-test-zookeeper-hl.helmreleasetest-67t9g.svc\n    zookeeper_auth_type: none\n    zookeeper_port: \"2181\"\n  repo: \"\"\nstatus: {}\n",
				"helmreleasetest-67t9g", namespace, -1)
			Expect(releaseCache.Manifest).To(Equal(manifest))

			Expect(releaseCache.ReleaseResourceMetas).To(Equal(getZookeeperDefaultReleaseResourceMeta(namespace, "zookeeper-test")))

			mataInfoParams := getZookeeperDefaultMetaInfoParams()
			Expect(releaseCache.MetaInfoValues).To(Equal(mataInfoParams))
		})
	})

})

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
			Kind: "ConfigMap",
			Namespace: namespace,
			Name: name + "-zookeeper-confd-conf",
		},
		{
			Kind: "ConfigMap",
			Namespace: namespace,
			Name: name + "-zookeeper-entrypoint",
		},
		{
			Kind: "Service",
			Namespace: namespace,
			Name: name + "-zookeeper-hl",
		},
		{
			Kind: "StatefulSet",
			Namespace: namespace,
			Name: name + "-zookeeper",
		},
		{
			Kind: "ReleaseConfig",
			Namespace: namespace,
			Name: name,
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
				Name: "zoo_cfg",
				Type: "kvPair",
				Value: "{\"autopurge.purgeInterval\":5,\"autopurge.snapRetainCount\":10,\"initLimit\":10,\"maxClientCnxns\":0,\"syncLimit\":5,\"tickTime\":9000}",
			},
			{
				Name: "zookeeper",
				Type: "kvPair",
				Value: "{\"zookeeper.client.port\":2181,\"zookeeper.jmxremote.port\":9911,\"zookeeper.leader.elect.port\":3888,\"zookeeper.peer.communicate.port\":2888}",
			},
		},
	}
}
