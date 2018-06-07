{
  local kube = self,

  matchTosVersion(config={}, version="")::
    if std.objectHas(config, "TosVersion") then
      if config.TosVersion == version then true
      else false
    else if version == "1.2" then true
    else false,

  tosVersionAdapter(config={}, tos1_2="", tos1_5="")::
    if kube.matchTosVersion(config, "1.2") then tos1_2
    else tos1_5,

  haAdapter(config={}, ha="", non_ha="")::
    if config.use_high_availablity then ha
    else non_ha,

  installName(name, config={})::
    if std.length(name) > 0 then name + config.Transwarp_Install_ID else '',

  tos:: {
    Base(wrapped, config):
      wrapped,
    ReplicationController(wrapped, config):
      local base = kube.tos.Base(wrapped, config);
      base {
        spec+: {
          [if $.tos.UserDefinedApplicationPause(config) then "replicas"]: 0,
        },
      },

    PdReplicationController(wrapped, config):
      $.tos.ReplicationController(wrapped, config),

    DaemonSet(wrapped, config):
      local base = kube.tos.Base(wrapped, config);
      local moduleName = base.spec.template.metadata._moduleName;
      // TODO fix pause daemonset with node selector
      base {
        spec+: {
          template+: $.tos.PodTemplateFinalizer(super.template, moduleName, config),
        },
      },

    Deployment(wrapped, config):
      local base = kube.tos.Base(wrapped, config);
      local moduleName = base.spec.template.metadata._moduleName;
      base {
        spec+: {
          [if $.tos.UserDefinedApplicationPause(config) then "replicas"]: 0,
          template+: $.tos.PodTemplateFinalizer(super.template, moduleName, config),
        },
      },

    StatefulSet(wrapped, config):
      local base = kube.tos.Base(wrapped, config);
      local moduleName = base.spec.template.metadata._moduleName;
      local generatedPvcs = kube.tos.UserDefinedPvcs(moduleName, config);
      base {
        spec+: {
          [if $.tos.UserDefinedApplicationPause(config) then "replicas"]: 0,
          template+: $.tos.PodTemplateFinalizer(super.template, moduleName, config),
          [if $.tos.Defined_Customized_Volumes(config) && std.length(generatedPvcs) == 0 then "volumeClaimTemplates"]:: [],
          [if $.tos.Defined_Customized_Volumes(config) && std.length(generatedPvcs) > 0 then "volumeClaimTemplates"]: generatedPvcs,
        },
      },

    PodTemplateFinalizer(template, moduleName, config):
      local generatedVolumes = kube.tos.UserDefinedVolumes(moduleName, config);
      local generatedAnnotations = kube.tosVersionAdapter(config, {}, kube.tos.UserDefinedAutoInjectedVolume(config));
      template {
        metadata+: {
          [if std.length(generatedAnnotations) > 0 then "annotations"]+: generatedAnnotations,
        },
        spec+: {
          containers: [$.tos.PodContainerFinalizer(c, c.name, moduleName, config) for c in super.containers],
          [if $.tos.Defined_Customized_Volumes(config) && std.length(generatedVolumes) == 0 then "volumes"]:: [],
          [if $.tos.Defined_Customized_Volumes(config) && std.length(generatedVolumes) > 0 then "volumes"]: generatedVolumes,
        },
      },

    PodContainerFinalizer(container, containerName, moduleName, config):
      local superResources = if std.objectHas(container, "resources") then container.resources else {};
      local finalResources = kube.tos.ContainerResourceFinalizer(superResources, containerName, moduleName, config);
      local generatedVolumeMounts = kube.tos.UserDefinedVolumeMounts(containerName, moduleName, config);
      container {
        resources: finalResources,
        [if $.tos.Defined_Customized_Volumes(config) && std.length(generatedVolumeMounts) == 0 then "volumeMounts"]:: [],
        [if $.tos.Defined_Customized_Volumes(config) && std.length(generatedVolumeMounts) > 0 then "volumeMounts"]: generatedVolumeMounts,
        [if std.objectHas(config, "Transwarp_Finalizer_Env_List") && std.length(config.Transwarp_Finalizer_Env_List) > 0 && std.objectHas(container, "env") then "env"]+: kube.tos.Generate_env_list(config.Transwarp_Finalizer_Env_List, []),
        // in the case container does not contain env list
        [if std.objectHas(config, "Transwarp_Finalizer_Env_List") && std.length(config.Transwarp_Finalizer_Env_List) > 0 && !std.objectHas(container, "env") then "env"]: kube.tos.Generate_env_list(config.Transwarp_Finalizer_Env_List, []),
      },

    ContainerResourceFinalizer(resource, containerName, moduleName, config):
      # when Transwarp_App_Resources is specified, we can override the requests and limits
      # when Transwarp_No_Limit is specified, we should remove the limits

      local userCpuLimit = kube.tos.UserDefinedResourceAmount("cpu", "limits", containerName, moduleName, config);
      local cpuLimit = if std.length(userCpuLimit) > 0 then userCpuLimit else (if std.objectHas(resource, "limits") && std.objectHas(resource.limits, "cpu") then resource.limits.cpu else "");
      local userCpuRequest = kube.tos.UserDefinedResourceAmount("cpu", "requests", containerName, moduleName, config);
      local cpuRequest = if std.length(userCpuRequest) > 0 then userCpuRequest else (if std.objectHas(resource, "requests") && std.objectHas(resource.requests, "cpu") then resource.requests.cpu else "");

      local userMemoryLimit = kube.tos.UserDefinedResourceAmount("memory", "limits", containerName, moduleName, config);
      local memoryLimit = if std.length(userMemoryLimit) > 0 then (std.toString(userMemoryLimit) + "Gi") else (if std.objectHas(resource, "limits") && std.objectHas(resource.limits, "memory") then resource.limits.memory else "");
      local userMemoryRequest = kube.tos.UserDefinedResourceAmount("memory", "requests", containerName, moduleName, config);
      local memoryRequest = if std.length(userMemoryRequest) > 0 then (std.toString(userMemoryRequest) + "Gi") else (if std.objectHas(resource, "requests") && std.objectHas(resource.requests, "memory") then resource.requests.memory else "");

      local showNoLimit = std.objectHas(config, "Transwarp_No_Limit") && config.Transwarp_No_Limit;
      {
        [if !showNoLimit then "limits"]: {
          [if std.length(memoryLimit) > 0 then "memory"]: memoryLimit,
          [if std.length(cpuLimit) > 0 then "cpu"]: cpuLimit,
        },
        requests: {
          [if std.length(memoryRequest) > 0 || (showNoLimit && std.length(memoryLimit) > 0) then "memory"]: if std.length(memoryRequest) > 0 then memoryRequest else memoryLimit,
          [if std.length(cpuRequest) > 0 || (showNoLimit && std.length(cpuLimit) > 0) then "cpu"]: if std.length(cpuRequest) > 0 then cpuRequest else cpuLimit,
        },
      },

    UserDefinedResourceAmount(resourceType, resourceCategory, containerName, moduleName, config):
      if std.objectHas(config, "Transwarp_App_Resources") &&
         std.objectHas(config.Transwarp_App_Resources, moduleName) &&
         std.objectHas(config.Transwarp_App_Resources[moduleName], containerName) &&
         std.objectHas(config.Transwarp_App_Resources[moduleName][containerName], resourceCategory) &&
         std.objectHas(config.Transwarp_App_Resources[moduleName][containerName][resourceCategory], resourceType) then
        std.toString(config.Transwarp_App_Resources[moduleName][containerName][resourceCategory][resourceType])
      else "",

    UserDefinedVolumes(moduleName, config):
      if $.tos.Defined_Customized_Volumes(config) then
        [
          (if v.type == "tosdisk" then
            kube.tos.UserDefinedTosDisk(v, config)
          else if v.type == "nfs" then
            // TODO fix this
            {}
          else if v.type == "hostsharedir" then
            kube.tos.UserDefinedHostShareDir(v, config)
          else if v.type == "hostpath" then
            kube.tos.UserDefinedHostPathVolume(v, config)
          else
            error ("Undefined volume type " + std.toString(v.type))
          )
          for v in config.Transwarp_Customized_Volumes if v.module == moduleName && v.type != "pvc"
        ] else [],

    UserDefinedVolumeMounts(containerName, moduleName, config):
      std.flattenArrays([
        ([{
          mountPath: m.path,
          name: v.volume_name,
          [if std.objectHas(m, "sub_path") && std.length(m.sub_path) > 0 then "subPath"]: m.sub_path,
          [if std.objectHas(m, "readOnly") && m.readOnly then "readOnly"]: true,
        }
          for m in v.mount_points if m.container_name == containerName])
        for v in config.Transwarp_Customized_Volumes if v.module == moduleName
      ]),

    VolumeMountsInContainer(container, moduleName, config)::
      local containerName = container.name;
      local generatedVolumeMounts = $.v1.VolumeMountsFromConfig(moduleName, containerName, config);
      // when volume config specifies, override the default volume mounts
      container {
        [if $.tos.Defined_Customized_Volumes(config) && std.length(generatedVolumeMounts) > 0 then "volumeMounts"]: generatedVolumeMounts,
        // when no volume mounts generated, hide the volume mounts
        [if $.tos.Defined_Customized_Volumes(config) && std.length(generatedVolumeMounts) == 0 then "volumeMounts"]:: [],
      },

    UserDefinedPvcs(moduleName, config):
      if $.tos.Defined_Customized_Volumes(config) then [
        $.tos.UserDefinedPersistentVolumeClaim(v, moduleName, config)
        for v in config.Transwarp_Customized_Volumes if v.type == "pvc" && v.module == moduleName
      ] else [],

    Defined_Customized_Volumes(config):
      std.objectHas(config, "Transwarp_Customized_Volumes") && std.length(config.Transwarp_Customized_Volumes) > 0,

    UserDefinedTosDisk(volume_config, config):
      local tosdisk_configs = {
        storageClass: volume_config.class_name,
        // TODO fix the iops_limit
        limits: {},
        size: volume_config.size,
        accessMode: "ReadWriteOnce",
      };
      kube.v1.TosDisk(volume_config.volume_name, tosdisk_configs),

    UserDefinedPersistentVolumeClaim(volume_config, moduleName, config):
      local pvc_configs = {
        storageClass: volume_config.class_name,
        // TODO fix the iops_limit
        limits: {},
        size: volume_config.size,
        accessModes: ["ReadWriteOnce"],
      };
      kube.v1.PersistentVolumeClaim(volume_config.volume_name, moduleName, pvc_configs, config),

    UserDefinedHostShareDir(volume_config, config):
      // TODO apply size & iops_limit
      local path = if std.objectHas(volume_config, "configs") && std.objectHas(volume_config.configs, "path") then volume_config.configs.path else error ("path is required in HostShareDir");
      local namespace = if std.objectHas(volume_config.configs, "namespace") then volume_config.configs.namespace else "";
      kube.v1.HostShareDirVolume(volume_config.volume_name, path, namespace),

    UserDefinedHostPathVolume(volume_config, config):
      local path = if std.objectHas(volume_config, "configs") && std.objectHas(volume_config.configs, "path") then volume_config.configs.path else error ("path is required in HostPathVolume");
      kube.v1.HostPath(volume_config.volume_name, path),

    UserDefinedApplicationPause(config):
      // Use Transwarp_Application_Pause to set replicas of all pod template to 0
      std.objectHas(config, "Transwarp_Application_Pause") && config.Transwarp_Application_Pause,

    UserDefinedAutoInjectedVolume(config):
      // This function help to generate the annotation of transwarp.io/tos-volume, which is used to auto-inject secret or configMap in pod
      // Expected input
      // Transwarp_Auto_Injected_Volumes = [{
      //   kind: required, Secret/ConfigMap
      //   volumeName: required
      //   name: optional, the name of secret / configMap
      //   selector: optional, the label selector for secret / configMap, either name or selector should be specified
      //   namespace: optional, the namespace of configMap. it is used only when kind == ConfigMap
      // }, ... ]
      local volume_statement = if std.objectHas(config, "Transwarp_Auto_Injected_Volumes") && std.length(config.Transwarp_Auto_Injected_Volumes) > 0 then
        [
          (if item.kind == "Secret" then kube.tos.UserDefinedAutoInjectedSecretVolume(item)
          else kube.tos.UserDefinedAutoInjectedConfigMapVolume(item))
          for item in config.Transwarp_Auto_Injected_Volumes
        ] else [];
      if std.length(volume_statement) > 0 then {
        "transwarp.io/tos-volume": std.toString(volume_statement),
      } else {},

    UserDefinedAutoInjectedSecretVolume(config):
      {
        kind: "Secret",
        [if std.objectHas(config, "name") && std.length(config.name) > 0 then "name"]: config.name,
        [if std.objectHas(config, "selector") && std.length(config.selector) > 0 then "selector"]: config.selector,
        volumeName: config.volumeName,
      },

    UserDefinedAutoInjectedConfigMapVolume(config):
      {
        kind: "ConfigMap",
        [if std.objectHas(config, "name") && std.length(config.name) > 0 then "name"]: config.name,
        [if std.objectHas(config, "selector") && std.length(config.selector) > 0 then "selector"]: config.selector,
        [if std.objectHas(config, "namespace") && std.length(config.namespace) > 0 then "namespace"]: config.namespace,
        volumeName: config.volumeName,
      },

    Generate_env_list(config, default)::
      # step 0: type check
      if std.type(config) != "array" then
        error ("std.filterMap first param must be array, got " + std.type(config))
      else if std.type(default) != "array" then
        error ("std.filterMap second param must be array, got " + std.type(default))
      else
        # step 1: merge & remove duplicates(judge by 'key')
        local config_not_contains(ele) = std.foldl(function(x, y) if y.key == ele.key then false else x, config, true);
        local ans = config + std.filter(config_not_contains, default);

        # step 2: ouput answer, map {key: "", value: ""} to {name: ""， value: ""}
        std.map(function(ele) { name: ele.key, value: ele.value }, ans),
  },

  v1:: {

    local ApiVersion = {
      apiVersion: "v1",
    },

    Metadata(name="", generateName="", config={}): {
      [if std.length(name) > 0 then "name"]: name,
      [if std.length(generateName) > 0 then "generateName"]: generateName,
      [if std.objectHas(config, "Customized_Namespace") && std.length(config.Customized_Namespace) > 0 then "namespace"]: config.Customized_Namespace,
    },

    ReplicationController(name="", generateName="", moduleName="", config={}): ApiVersion {
      local _moduleName = if std.length(moduleName) > 0 then moduleName else if std.length(name) > 0 then name else generateName,
      local _name = if std.length(name) > 0 then name + config.Transwarp_Install_ID else '',

      kind: "ReplicationController",
      metadata: $.v1.Metadata(_name, generateName, config) {
        labels: $.ReservedLabels(_moduleName, config),
        # keep this in case in future, we need to add a common annotation
        annotations: {},
      },

      spec: {
        selector: $.ReservedSelector(_moduleName, config),
        template: $.v1.PodTemplate(_moduleName, config),
      },
    },

    PdReplicationController(name="", generateName="", moduleName="", pdContainerName="pd", config={}): $.v1.ReplicationController(name, generateName, moduleName, config) {
      local _moduleName = if std.length(moduleName) > 0 then moduleName else if std.length(name) > 0 then name else generateName,
      #            metadata+: {
      #                labels+: {
      #                    [moduleName]: "1"
      #                },
      #            },
      spec+: {
        template+: {
          metadata+: {
            annotations:: super.annotations,
            labels+: {
              #                           [_moduleName + ".install." + config.Transwarp_Install_ID]: "true",
              #                           [_moduleName]: "1",
              "transwarp.pd.pod": "true",
            },
          },
          spec+: {
            containers: [
              $.v1.PodContainer(pdContainerName) {
                args: [
                  "ls",
                ],
                image: std.toString(config.Transwarp_Registry_Server) + "/jenkins/transwarppd:live",
                imagePullPolicy:: super.imagePullPolicy,
              },
            ],
            podDiskSpec: {
              isPersistentDirPod: true,
            },
            restartPolicy: "OnFailure",
          },
        },
      },

    },

    Service(name="", generateName="", moduleName="", selectorModuleName="", config={}): ApiVersion {
      local _moduleName = if std.length(moduleName) > 0 then moduleName else if std.length(name) > 0 then name else generateName,
      local _name = if std.length(name) > 0 then name + config.Transwarp_Install_ID else '',
      local _selectorModuleName = if std.length(selectorModuleName) > 0 then selectorModuleName else _moduleName,

      kind: "Service",
      metadata: $.v1.Metadata(_name, generateName, config) {
        labels: $.ReservedLabels(_moduleName, config) + {
          "k8s-app": _moduleName,
        },
        annotations: {},
      },
      spec: {
        selector: $.ReservedSelector(_selectorModuleName, config),
      },
    },

    NodePortService(name="", generateName="", moduleName="", selectorModuleName="", config={}): $.v1.Service(name, generateName, moduleName, selectorModuleName, config) {
      metadata+: {
        labels+: {
          "kubernetes.io/cluster-service": "true",
        },
      },

      spec+: {
        type: "NodePort",
      },
    },

    HeadlessService(name="", generateName="", moduleName="", selectorModuleName="", config={}): $.v1.Service(name, generateName, moduleName, selectorModuleName, config) {
      metadata+: {
        labels+: {
          "kubernetes.io/headless-service": "true",
        },
      },
      spec+: {
        clusterIP: "None",
      },
    },


    DummyService(providesInfo={}, config={}): $.v1.Service(generateName="app-dummy-", moduleName="dummy", config=config) {
      local metaAnnotation = {
        [if std.length(providesInfo) > 0 then "provides"]: providesInfo,
      },
      local _app_labels = if std.objectHasAll(config, "Transwarp_App_Labels") && std.length(config.Transwarp_App_Labels) > 0 then config.Transwarp_App_Labels else {},

      metadata+: {
        annotations+: {
          [if std.length(metaAnnotation) > 0 then "transwarp.meta"]: std.toString(metaAnnotation),
        },
        labels: $.ReservedLabels("app-dummy", config) + {
          "transwarp.meta": "true",
          "transwarp.svc.scope": "app",
          [if std.objectHasAll(config, "Transwarp_App_Scope") && std.length(config.Transwarp_App_Scope) > 0 then "transwarp.scope"]: config.Transwarp_App_Scope,
        } + _app_labels,
      },
      spec: {
        clusterIP: "None",
        ports: [{
          port: 5000,
          targetPort: 5000,
        }],
        selector: {},
      },
    },

    PodTemplate(moduleName, config): {
      metadata: $.v1.Metadata() {
        _moduleName:: moduleName,
        annotations: $.ReservedPodAnnotations(config),
        labels: $.ReservedLabels(moduleName, config) {
          [if std.objectHasAll(config, "Transwarp_App_Meta") && std.length(config.Transwarp_App_Meta) > 0 && std.objectHasAll(config.Transwarp_App_Meta, "name") then "transwarp.meta.app.name"]: config.Transwarp_App_Meta.name,
          [if std.objectHasAll(config, "Transwarp_App_Meta") && std.length(config.Transwarp_App_Meta) > 0 && std.objectHasAll(config.Transwarp_App_Meta, "version") then "transwarp.meta.app.version"]: config.Transwarp_App_Meta.version,
          [if std.objectHasAll(config, "Transwarp_App_Meta") && std.length(config.Transwarp_App_Meta) > 0 && std.objectHasAll(config.Transwarp_App_Meta, "id") then "transwarp.meta.app.id"]: std.toString(config.Transwarp_App_Meta.id),
        },
      },
      spec: {
      },
    },

    PodContainer(name): {
      imagePullPolicy: "Always",
      name: name,
    },

    PersistentDirVolume(name, selector): {
      name: name,
      persistentDir: {
        podSelector: selector,
      },
    },

    PersistentVolumeClaim(name, moduleName, storageConfig={}, config={}): {
      metadata: $.v1.Metadata(name) {
        labels: $.ReservedLabels(moduleName, config),
        annotations: {
          "volume.beta.kubernetes.io/storage-class": storageConfig.storageClass,
        },
      },
      spec: {
        accessModes: storageConfig.accessModes,
        resources: {
          requests: {
            storage: storageConfig.size,
          },
          [if std.length(storageConfig.limits) > 0 then "limits"]: storageConfig.limits,
        },
      },
    },

    EmptyVolVolume(name, size): {
      name: name,
      emptyVol: {
        mount: {
          size: std.toString(size),
        },
      },
    },

    HostPath(name, path): {
      name: name,
      hostPath: {
        path: path,
      },
    },

    TosDisk(name, storageConfig): {
      name: name,
      tosDisk: {
        name: name,
        storageType: storageConfig.storageClass,
        capability: storageConfig.size,
        accessMode: storageConfig.accessMode,
      },
    },

    HostShareDirVolume(name, path, namespace=""): {
      name: name,
      hostShareDir: {
        path: path,
        [if std.length(namespace) != 0 then "namespace"]: namespace,
      },
    },
    ContainerResources(cpu_request=0, memory_limit=0, cpu_limit=0, memory_request=0): {
      local _cpu_request = if std.type(cpu_request) == "object" && std.objectHas(cpu_request, "request") then cpu_request.request else cpu_request,
      local _cpu_limit = if std.type(cpu_request) == "object" && std.objectHas(cpu_request, "limit") then cpu_request.limit else cpu_limit,
      local _memory_request = if std.type(memory_limit) == "object" && std.objectHas(memory_limit, "request") then memory_limit.request else memory_request,
      local _memory_limit = if std.type(memory_limit) == "object" && std.objectHas(memory_limit, "limit") then memory_limit.limit else memory_limit,
      limits: {
        [if _memory_limit > 0 then "memory"]: std.toString(_memory_limit) + "Gi",
        [if _cpu_limit > 0 then "cpu"]: std.toString(_cpu_limit),
      },
      requests: {
        [if _memory_request > 0 then "memory"]: std.toString(_memory_request) + "Gi",
        [if _cpu_request > 0 then "cpu"]: std.toString(_cpu_request),
      },
    },

    EnvFieldPath(name, path): {
      name: name,
      valueFrom: {
        fieldRef: {
          fieldPath: path,
        },
      },
    },
  },

  "apps/v1beta1":: {
    local ApiVersion = {
      apiVersion: "apps/v1beta1",
    },

    StatefulSet(name="", generateName="", moduleName="", config={}): ApiVersion {
      local _moduleName = if std.length(moduleName) > 0 then moduleName else if std.length(name) > 0 then name else generateName,
      local _installName = kube.installName(name, config),
      kind: "StatefulSet",
      metadata: $.v1.Metadata(_installName, generateName, config) {
        annotations: {},
        labels: $.ReservedLabels(_moduleName, config),
      },
      spec: {
        serviceName: _installName,
        selector: {
          matchLabels: $.ReservedSelector(_moduleName, config),
        },
        template: $.v1.PodTemplate(_moduleName, config),
      },
    },
  },

  "extensions/v1beta1":: {
    local ApiVersion = {
      apiVersion: "extensions/v1beta1",
    },

    Deployment(name="", generateName="", moduleName="", config={}): ApiVersion {
      local _moduleName = if std.length(moduleName) > 0 then moduleName else if std.length(name) > 0 then name else generateName,
      local _name = if std.length(name) > 0 then name + config.Transwarp_Install_ID else '',
      kind: "Deployment",
      metadata: $.v1.Metadata(_name, generateName, config) {
        annotations: {},
        labels: $.ReservedLabels(_moduleName, config),
      },
      spec: {
        selector: {
          matchLabels: $.ReservedSelector(_moduleName, config),
        },
        template: $.v1.PodTemplate(_moduleName, config),
      },
    },

    DeploymentStrategy(config={}):
      if std.type(config) != "object" then
        error ("config must be an object")
      else if std.length(config) == 0 then
        {}
      else if !std.objectHas(config, "type") then
        error ("config does not have type attribute")
      else if config.type == "Recreate" then
        {
          type: "Recreate",
        }
      else if config.type == "RollingUpdate" then
        {
          type: "RollingUpdate",
          rollingUpdate: {
            maxUnavailable: config.rolling_update_configs.max_unavailable,
            maxSurge: config.rolling_update_configs.max_surge,
          },
        }
      else
        error ("invalid configs for Deployment Strategy "),


    DaemonSet(name="", generateName="", moduleName="", config={}): ApiVersion {
      local _moduleName = if std.length(moduleName) > 0 then moduleName else if std.length(name) > 0 then name else generateName,
      local _name = if std.length(name) > 0 then name + config.Transwarp_Install_ID else '',
      kind: "DaemonSet",
      metadata: $.v1.Metadata(_name, generateName, config) {
        labels: $.ReservedLabels(_moduleName, config),
      },
      spec: {
        selector: {
          matchLabels: $.ReservedSelector(_moduleName, config),
        },
        template: $.v1.PodTemplate(_moduleName, config),
      },
    },

  },

  instance_selector(config)::
    if std.objectHasAll(config, "Customized_Instance_Selector") && std.length(config.Customized_Instance_Selector) > 0 then
      config.Customized_Instance_Selector
    else {
      "transwarp.install": config.Transwarp_Install_ID,
    },

  ReservedLabels(moduleName, config):: {
    "transwarp.name": moduleName,
    [if std.objectHas(config, "Transwarp_Alias") && std.length(config.Transwarp_Alias) > 0 then "transwarp.alias"]: config.Transwarp_Alias,
  } + $.instance_selector(config),


  ReservedSelector(moduleName, config):: {
    "transwarp.name": moduleName,
  } + $.instance_selector(config),

  ReservedPodAnnotations(config):: {
    [if std.objectHas(config, "Transwarp_App_Name") && std.length(config.Transwarp_App_Name) > 0 then "transwarp.app"]: config.Transwarp_App_Name,
    [if std.objectHas(config, "Transwarp_Namespace_Owner") && std.length(config.Transwarp_Namespace_Owner) > 0 then "transwarp.namespace.owner"]: config.Transwarp_Namespace_Owner,
    [if kube.matchTosVersion(config, "1.5") then "cni.networks"]: if std.objectHas(config, "Transwarp_Cni_Network") && std.length(config.Transwarp_Cni_Network) > 0 then config.Transwarp_Cni_Network else "overlay",
  } + (if kube.matchTosVersion(config, "1.5") && std.objectHas(config, "Transwarp_Network_Limit") then {
    [if std.objectHas(config.Transwarp_Network_Limit, "ingress") then "tos.network.ingress_limit"]: std.toString(config.Transwarp_Network_Limit.ingress),
    [if std.objectHas(config.Transwarp_Network_Limit, "egress") then "tos.network.egress_limit"]: std.toString(config.Transwarp_Network_Limit.egress),
  } else {}) + (if std.objectHas(config, "Transwarp_Metrics_Scrape") then {
    "prometheus.io/scrape": "true",
    "prometheus.io/port": if std.objectHas(config.Transwarp_Metrics_Scrape, "port") then std.toString(config.Transwarp_Metrics_Scrape.port) else "5556",
    "prometheus.io/path": if std.objectHas(config.Transwarp_Metrics_Scrape, "path") then std.toString(config.Transwarp_Metrics_Scrape.path) else "/metrics",
  } else {}),

  NameSpace(config)::
    if std.objectHas(config, "Transwarp_Install_Namespace") && std.length(config.Transwarp_Install_Namespace) > 0 then config.Transwarp_Install_Namespace else "default",

  NSFromSvcMeta(obj)::
    if std.objectHasAll(obj, "metadata") &&
       std.objectHasAll(obj.metadata, "namespace") then
      obj.metadata.namespace
    else "",

  AddrFromRC(rc, suffix="")::
    if std.length(rc) > 0 then (
      if !std.objectHas(rc, "metadata") then
        error ("rc does not have metadata")
      else if !std.objectHas(rc, "spec") then
        error ("rc does not have spec")
      else
        kube.DNSFromRCMeta(rc.metadata, rc.spec.replicas, suffix)
    ) else "",

  AddrFromStatefulSet(statefulSet, suffix="")::
    if std.length(statefulSet) > 0 then (
      if !std.objectHas(statefulSet, "metadata") then
        error ("statefulset does not have metadata")
      else if !std.objectHas(statefulSet, "spec") then
        error ("statefulset does not have spec")
      else
        kube.DNSFromStatefulSetMeta(statefulSet.metadata, statefulSet.spec.replicas, suffix)
    ) else "",

  DNSFromPrefixedId(prefix, config, suffix="svc")::
    std.join(".", [std.toString(prefix) + std.toString(config.Transwarp_Install_ID), std.toString(config.Transwarp_Install_Namespace), if std.length(suffix) > 0 then suffix]),

  DNSFromPod(prefix, config, suffix="")::
    std.join(".", [std.toString(prefix) + std.toString(config.Transwarp_Install_ID), "pod", kube.NameSpace(config), if std.length(suffix) > 0 then suffix]),

  DNSFromRCMeta(rcMeta, replicas, suffix="")::
    if std.type(rcMeta) == "object" && std.length(rcMeta) > 0 then (
      if std.length(rcMeta.name) == 0 || std.length(rcMeta.namespace) == 0 then
        error ("meta does not have name or namespace")
      else
        std.join(",", std.makeArray(replicas, function(i)
          std.join(".", [std.toString(i), rcMeta.name, rcMeta.namespace, "rc", if std.length(suffix) > 0 then suffix])
        ))
    ) else "",

  DNSFromStatefulSetPod(name, id, config={}, suffix="")::
    std.join(".", [kube.installName(name, config) + "-" + std.toString(id), std.toString(config.Transwarp_Install_Namespace), "pod", if std.length(suffix) > 0 then suffix]),

  DNSFromStatefulSetMeta(statefulSetMeta, replicas, suffix="")::
    if std.type(statefulSetMeta) == "object" && std.length(statefulSetMeta) > 0 then (
      local namespace = if std.length(statefulSetMeta.namespace) == 0 then "default" else statefulSetMeta.namespace;
      if std.length(statefulSetMeta.name) == 0 then
        error ("meta does not have name")
      else
        std.join(",", std.makeArray(replicas, function(i)
          std.join(".", [std.join("-", [statefulSetMeta.name, std.toString(i)]), namespace, "pod", if std.length(suffix) > 0 then suffix])))
    ) else "",

  DNSFromNodeSvcMeta(obj={}, ns="namespace", meta="metadata", anno="annotations", name_list="")::
    if std.objectHasAll(obj, meta) &&
       std.objectHasAll(obj[meta], ns) &&
       std.length(obj[meta][ns]) > 0 &&
       std.objectHasAll(obj[meta], anno) &&
       std.objectHasAll(obj[meta][anno], name_list) then
      local nameArr = std.split(obj[meta][anno][name_list], ",");
      local _nl = std.filter(function(x) std.length(x) > 0, nameArr);

      local _ns = obj[meta][ns];

      std.join(",", std.makeArray(std.length(_nl), function(i)
        std.join(".", [_nl[i], _ns, "svc"])
      ))

    else "",

  DNSFromSvcMeta(config, fieldName)::
    if std.objectHas(config, fieldName) && std.length(config[fieldName]) > 0 then (
      if std.length(config[fieldName].name) == 0 || std.length(config[fieldName].namespace) == 0 then
        error ("meta does not have name or namespace")
      else
        config[fieldName].name + "." + config[fieldName].namespace + ".svc"
    ) else "",

  AffinityAnnotations(nodeAffinity={}, podAffinity={}, podAntiAffinity={})::
    {
      "scheduler.alpha.kubernetes.io/affinity": std.toString({
        [if std.length(nodeAffinity) > 0 then "nodeAffinity"]: nodeAffinity,
        [if std.length(podAffinity) > 0 then "podAffinity"]: podAffinity,
        [if std.length(podAntiAffinity) > 0 then "podAntiAffinity"]: podAntiAffinity,
      }),
    },

  NodeAntiAffinityAnnotations(config, moduleName="", nodeAffinity={})::
    local annotations = $.ReservedLabels(moduleName, config);
    local res =
      {
        requiredDuringSchedulingIgnoredDuringExecution:
          [{
            labelSelector: {
              matchLabels: annotations,
            },
            namespaces: [kube.NameSpace(config)],
            topologyKey: "kubernetes.io/hostname",
          }],
      };
    $.AffinityAnnotations(nodeAffinity=nodeAffinity, podAffinity={}, podAntiAffinity=res),

  CustomizedNodeAntiAffinityAnnotations(config, additional_labels)::
    local res =
      {
        requiredDuringSchedulingIgnoredDuringExecution:
          [{
            labelSelector: {
              matchLabels: additional_labels,
            },
            namespaces: [kube.NameSpace(config)],
            topologyKey: "kubernetes.io/hostname",
          }],
      };
    $.AffinityAnnotations(nodeAffinity={}, podAffinity={}, podAntiAffinity=res),
}