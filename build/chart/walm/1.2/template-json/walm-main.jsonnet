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
           value: overall_config.Advance_Config.port,
         },
         {
           name: "WALM_DEBUG",
           value: overall_config.Advance_Config.debug,
         },
         {
           name: "WALM_DB_NAME",
           value: overall_config.Advance_Config.dbname,
         },]+
         if std.objectHas(overall_config.MYSQL_CLIENT_CONFIG,"username") then
         [{ 
           name: "WALM_DB_USER",
           value: overall_config.MYSQL_CLIENT_CONFIG.username,
         },]else[]+
         if std.objectHas(overall_config.MYSQL_CLIENT_CONFIG,"password") then
         [{
           name: "WALM_DB_PASS",
           value: overall_config.MYSQL_CLIENT_CONFIG.password,
         },]else[]+
         [{
           name: "WALM_DB_TYPE",
           value: overall_config.Advance_Config.dbtype,
         },]+
         if std.objectHas(overall_config.MYSQL_CLIENT_CONFIG,"mysqladdresses") then
         [{
           name: "WALM_DB_HOST",
           value: overall_config.MYSQL_CLIENT_CONFIG.mysqladdresses+":"+overall_config.MYSQL_CLIENT_CONFIG.mysqlport,
         },]else[]+
         [{
           name: "WALM_TABLE_PREFIX",
           value: overall_config.Advance_Config.dbtabpre,
         },
         {
           name: "HTTP_READ_TIMEOUT",
           value: overall_config.Advance_Config.httpreadtimeout,
         },
         {
           name: "HTTP_READ_TIMEOUT",
           value: overall_config.Advance_Config.httpwritetimeout,
         },
         {
           name: "ZIPKIN_URL",
           value: overall_config.Advance_Config.zipkinurl,
         },
         {
           name: "TILLER_CONN_TIMEOUT",
           value: overall_config.Advance_Config.tiller_connection_timeout,
         },
         {
           name: "KUBE_CONTEXT",
           value: overall_config.Advance_Config.kube_context,
         },
         {
           name: "WALM_OAUTH",
           value: overall_config.Advance_Config.oauth,
         },
         {
           name: "WALM_JWTSECRET",
           value: overall_config.Advance_Config.JwtSecret,
         },];
  {

    'walm-hl-svc.json':
      walm.walm.headless_svc('walm-srv', overall_config),
    
    'walm-svc.json':
      walm.walm.svc('walm-srv', overall_config),

    'walm-deployment.json':
      walm.walm.deployment('walm-srv', overall_config , env),

    'dummy-svc.json':
      kube.v1.DummyService(providesInfo={
        WALM_CONFIG: {
          immediate_value: {
            walm_http_port: overall_config.Advance_Config.port,
          },
        },
      }, config=overall_config),
  }