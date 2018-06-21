{
  LICENSE_ADDRESS: '172.16.1.41:2181',
  Transwarp_Alias: 'walm_0000',
  Transwarp_App_Labels: {
    'transwarp.app': 'walm',
  },
  Transwarp_App_Name: 'walm',
  Transwarp_App_Scope: 'namespace',
  Transwarp_DEBUG: true,
  Transwarp_Install_ID: '1160',
  Transwarp_Install_Namespace: 'u5003',
  Transwarp_Namespace_Owner: 'flannel',
  Transwarp_Principal_Suffix: 'transwarp.local',
  Transwarp_Realm: 'TOS',
  Transwarp_Registry_Server: '172.16.1.41:5000',
  TosVersion: '1.5',

  App: {
    image: "172.16.1.99:5000/gold/walm:tdc-1.0",
    priority: 0,
    replicas: 1,
    resources: {
      cpu_limit: 4,
      cpu_request: 1,
      memory_limit: 4,
      memory_request: 2,
    }
  },

  
  Advance_Config: {
      debug: false,
      port: 8000,
      httpreadtimeout: 0,
      httpwritetimeout: 0,
      zipkinurl: "",
      tiller_connection_timeout: 0,
      kube_context: "default",
      oauth: false,  
      JwtSecret:"",
      
      dbname: "walm",
      dbtype: "mysql",
      dbtabpre: "",

    },

    MYSQL_CLIENT_CONFIG: {}, 
}
