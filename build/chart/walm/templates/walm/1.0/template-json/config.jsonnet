{
  LICENSE_ADDRESS: "",
  Transwarp_Alias: "",
  Transwarp_App_Name: "",
  Transwarp_Install_ID: "",
  Transwarp_Guardian_Address: "",
  Transwarp_Guardian_LDAP_Server_Address: "",
  Transwarp_App_Scope: "",
  Transwarp_App_Labels: {},
  Transwarp_DEBUG: true,
  Transwarp_Install_Namespace: "",
  Transwarp_Namespace_Owner: "",
  Transwarp_Principal_Suffix: "",
  Transwarp_Realm: "",
  Transwarp_Registry_Server: "",

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
    db_conf: {
      dbname: "walm",
      dbtype: "mysql",
      dbuser: "root",
      dbpass: "passwd",
      dbhost: "",
      dbtabpre: "",
    },
    conf: {
      debug: false,
      port: 8000,
      httpreadtimeout: 0,
      httpwritetimeout: 0,
      zipkinurl: "",
      tiller_connection_timeout: 0,
      kube_context: "default",
      oauth: false,  
      JwtSecret:"",
    },
  },
  
  # depended applications
  txsql_rc: {},
  # TosVersion: "1.5",
}
