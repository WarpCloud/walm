// Copyright 2016 Transwarp Inc. All rights reserved.

local app = import '../../../applib/app.libsonnet';
local kube = import '../../../applib/kube.libsonnet';

{
  common:: {
    configs:: {
      shared_kv(config): {
        txsql: {
          txsqlnodes: std.join(
            ',', std.makeArray(
              config.App.txsql.replicas, function(i)
                std.join('.', [
                  'txsql-' + config.Transwarp_Install_ID + '-' + std.toString(i),
                  config.Transwarp_Install_Namespace,
                  'pod',
                ])
            )
          ),
        },
      },
      shared_env(config): {
      },
    },

    configMap:: {
      volumeMounts(config):: [
        { name: 'txsql-entrypoint', mountPath: '/boot' },
        { name: 'txsql-confd-conf', mountPath: '/etc/confd' },
      ],

      volumes(config):: [
        {
          name: 'txsql-entrypoint',
          configMap: {
            name: 'txsql-entrypoint-' + config.Transwarp_Install_ID,
            items: [
              { key: 'entrypoint.sh', path: 'entrypoint.sh', mode: 493 },
            ],
          },
        },
        {
          name: 'txsql-confd-conf',
          configMap: {
            name: 'txsql-confd-conf-' + config.Transwarp_Install_ID,
            items: [
              { key: 'txsql.toml', path: 'conf.d/txsql.toml' },
              { key: 'txsql-confd.conf', path: 'txsql-confd.conf' },
              { key: 'db.properties.tmpl', path: 'templates/db.properties.tmpl' },
              { key: 'install_conf.sh.tmpl', path: 'templates/install_conf.sh.tmpl' },
            ],
          },
        },
      ],
    },
  },

  txsql:: {
    statefulset(_name, config, env)::
      kube.tos.StatefulSet(
        kube['apps/v1beta1'].StatefulSet(name=_name + '-', moduleName=_name, config=config) {
          spec+: {
            replicas: config.App.txsql.replicas,
            template+: {
              metadata+: {
                annotations+: { configmap_md5: config.configmap_md5 } + kube.NodeAntiAffinityAnnotations(config, moduleName=_name),
              },
              spec+: {
                terminationGracePeriodSeconds: 0,
                hostNetwork: config.App.txsql.use_host_network,
                priority: config.App.txsql.priority,
                containers: [
                  kube.v1.PodContainer(_name) {
                    image: config.App.txsql.image,
                    args: ['/boot/entrypoint.sh'],
                    env: env + app.rc_env(config.App.txsql.env_list, []),
                    resources: kube.v1.ContainerResourcesV2(config.App.txsql.resources),
                    readinessProbe: {
                      exec: {
                        command: [
                          '/bin/bash',
                          '-c',
                          'ulimit -c 0; cd /usr/bin/txsql/tools && ./txsql.sh status',
                        ],
                      },
                      periodSeconds: 60,
                      initialDelaySeconds: 240,
                    },
                    volumeMounts: [
                      { mountPath: '/var/txsqldata', name: 'txsql-data', subPath: 'data' },
                      { mountPath: '/etc/conf/txsql', name: 'txsql-data', subPath: 'etc' },
                      { mountPath: '/var/txsqllog', name: 'txsql-log' },
                    ] + $.common.configMap.volumeMounts(config),
                  },
                ],
                volumes: $.common.configMap.volumes(config),
              },
            },
            volumeClaimTemplates: [
              kube.v1.PersistentVolumeClaim(name='txsql-data',
                                            moduleName='txsql-data',
                                            storageConfig=config.App.txsql.resources.storage.data,
                                            config=config),
              kube.v1.PersistentVolumeClaim(name='txsql-log',
                                            moduleName='txsql-log',
                                            storageConfig=config.App.txsql.resources.storage.log,
                                            config=config),
            ],
          },
        },
        config
      ),

    headless_svc(_name, config)::
      kube.v1.HeadlessService(name=_name + '-hl-', moduleName=_name + '-hl', selectorModuleName=_name, config=config) {
        metadata+: {
          annotations+: {
            'service.alpha.kubernetes.io/tolerate-unready-endpoints': 'true',
          },
        },
        spec+: {
          ports: [
            { name: 'mysql-port', port: 3306, protocol: 'TCP' },
          ],
        },
      },

    svc(_name, config)::
      kube.v1.NodePortService(name=_name + '-', moduleName=_name, config=config) {
        spec+: {
          ports: [
            { name: 'mysql-port', port: 3306, protocol: 'TCP', targetPort: 3306 },
          ],
        },
      },
  },
}
