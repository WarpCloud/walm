// Copyright 2016 Transwarp Inc. All rights reserved.

// import application function library
local app = import '../../../applib/app.libsonnet';
local kube = import '../../../applib/kube.libsonnet';

local txsql = import './txsql.jsonnet';

// user-defined data
local default_config = import 'config.jsonnet';

local configmapfiles = {
  'txsql.toml': importstr 'files/txsql.toml',
  'db.properties.tmpl': importstr 'files/db.properties.tmpl',
  'install_conf.sh.tmpl': importstr 'files/install_conf.sh.tmpl',
};
local configmapfiles_md5 = std.md5(std.toString(configmapfiles));

function(config={})
  local overall_config = std.mergePatch(default_config, config);
  local confd_config = std.mergePatch(
    std.mergePatch(txsql.common.configs.shared_kv(overall_config), overall_config.Transwarp_Config),
    overall_config.Advance_Config
  );
  local configmap = configmapfiles {
    'txsql-confd.conf': std.manifestJsonEx(confd_config, '  '),
  };
  local configmap_md5 = configmapfiles_md5 + std.md5(std.toString(confd_config));
  {
    'txsql-entrypoint-configmap.json':
      kube.v1.ConfigMap(name='txsql-entrypoint-', moduleName='txsql-entrypoint', config=overall_config) {
        data: {
          'entrypoint.sh': importstr 'files/entrypoint.sh',
        },
      },

    'txsql-confd-conf-configmap.json':
      kube.v1.ConfigMap(name='txsql-confd-conf-', moduleName='txsql-confd-conf', config=overall_config) {
        data: configmap,
      },

    'txsql-statefulset.json':
      local env = [
        { name: 'SMOKE_TEST', value: 'Y' },
        { name: 'CONF_DIR', value: '/etc/conf/txsql' },
      ];
      txsql.txsql.statefulset('txsql', overall_config { configmap_md5: configmap_md5 }, env),

    'txsql-svc.json':
      txsql.txsql.svc('txsql', overall_config),

    'txsql-hl-svc.json':
      txsql.txsql.headless_svc('txsql', overall_config),

    'dummy-svc.json':
      kube.v1.DummyService(providesInfo={
        MYSQL_CLIENT_CONFIG: {
          immediate_value: {
            type: 'txsql',
            mysqladdresses: txsql.common.configs.shared_kv(overall_config).txsql.txsqlnodes,
            mysqlport: overall_config.Advance_Config.install_conf.SQL_RW_PORT,
            username: overall_config.Advance_Config.db_properties['db.user'],
            password: overall_config.Advance_Config.db_properties['db.password'],
          },
        },
      }, config=overall_config),
  }
