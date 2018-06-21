local app = import '../../../applib/app.libsonnet';
local kube = import '../../../applib/kube.libsonnet';


local walm = import './walm.jsonnet';

// user-defined data
local default_config = import './config.jsonnet';

function(config={})
  local overall_config = std.mergePatch(default_config, config);
  local env = [
         {
           name: "WALM_HTTP_PORT",
           value: overall_config.Advance_Config.conf.port,
         },
         {
           name: "WALM_DEBUG",
           value: overall_config.Advance_Config.conf.debug,
         },
         {
           name: "WALM_DB_NAME",
           value: overall_config.Advance_Config.db_conf.dbname,
         },
         {
           name: "WALM_DB_PASS",
           value: overall_config.Advance_Config.db_conf.dbpass,
         },
         {
           name: "WALM_DB_TYPE",
           value: overall_config.Advance_Config.db_conf.dbtype,
         },
         {
           name: "WALM_DB_HOST",
           value: overall_config.Advance_Config.db_conf.dbhost,
         },
         {
           name: "WALM_TABLE_PREFIX",
           value: overall_config.Advance_Config.db_conf.dbtabpre,
         },
         {
           name: "HTTP_READ_TIMEOUT",
           value: overall_config.Advance_Config.conf.httpreadtimeout,
         },
         {
           name: "HTTP_READ_TIMEOUT",
           value: overall_config.Advance_Config.conf.httpwritetimeout,
         },
         {
           name: "ZIPKIN_URL",
           value: overall_config.Advance_Config.conf.zipkinurl,
         },
         {
           name: "TILLER_CONN_TIMEOUT",
           value: overall_config.Advance_Config.conf.tiller_connection_timeout,
         },
         {
           name: "KUBE_CONTEXT",
           value: overall_config.Advance_Config.conf.kube_context,
         },
         {
           name: "WALM_OAUTH",
           value: overall_config.Advance_Config.conf.oauth,
         },
         {
           name: "WALM_JWTSECRET",
           value: overall_config.Advance_Config.conf.JwtSecret,
         },
      ];
  {

    'walm-hl-svc.json':
      walm.walm.headless_svc('walm-srv', overall_config),
    
    'walm-svc.json':
      walm.walm.svc('walm-srv', overall_config),

    'walm-deployment.json':
      walm.walm.deployment('walm-srv', overall_config , env),

    'dummy-svc.json':
       kube.v1.DummyService(providesInfo=kube.tosVersionAdapter(overall_config, {
            WALM_ANNOTATION: {
                selector: {"transwarp.name": "walm",},
                resource_type: "deployments",
                key: "walm.annotations",
                },
            WALM_HL_SVC: {
                selector: {"transwarp.name": "walm-hl",},
                resource_type: "services",
                key: "",
                },
            }), config=overall_config)
  }
