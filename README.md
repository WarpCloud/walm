[1]: resource/Walm_Arch.jpg
[3]: resource/walm_logo.png

# Walm
![logo][3]

Walm is a micro service, based on Helm, that supports both Rest Api and Cli to manage the lifecycle of container based applications including those with dependencies.

Walm dynamically manages the dependencies of an application. An application can depend on the applications already existed, and the configurations of applications depending on would be injected automatically. Besides, once the configurations of applications depending on changes, the configurations would be injected again in real-time.

Walm supports more advanced Chart that use jsonnet as template engine to render kubernetes objects. It is more suitable to orchestrate and deploy complex applications, such as Big Data applications.

Walm supports finely grained authentication and authorization, that would make one user only have relevant authorization under kubernetes namespace scope.

Walm uses distributed event system to synchronize the application’s status in real-time.

## Architecture
![arch][1]

## Advantage
- Walm supports rest api to manage the lifecycle of applications.
- Walm supports the orchestration and deployment of complex applications.
- Walm supports the dynamic dependencies management.
- Walm supports the real-time synchronization of the application's status.
- Walm supports the finely grained authentication and authorization.
- Walm supports to retrieve the more detailed specification and status of applications.

## Features
- [Application Management](docs/application-management.md)
- [Application Groups Management](docs/application-groups-management.md)
- [Helm Charts Management](docs/helm-charts-management.md)
- [Kubernetes Resource Management](docs/kubernetes-resource-management.md)
- [High Availability](docs/high-availability.md)
- [Security](docs/security.md)

## Deploy
- [Run Walm On Linux Clusters](docs/run-walm-on-linux-clusters.md)
- [Run Walm On Google Kubernetes Engine Clusters](docs/run-walm-on-google-kubernetes-engine-clusters.md)
## Get Started
#### Deploy walm on kubernetes cluster
- [Run Walm On Linux Clusters](docs/run-walm-on-linux-clusters.md)
- [Run Walm On Google Kubernetes Engine Clusters](docs/run-walm-on-google-kubernetes-engine-clusters.md)
#### Install Helm
If helm is not installed, download executable file [here](https://github.com/WarpCloud/helm/releases), move to /usr/local/bin

#### Install zookeeper && kafka
Get started to install products when succeed to deploy walm on your kubernetes clusters.<br>
1. Visit https://github.com/WarpCloud/walm-charts, get kafka-6.1.0.tgz, zookeeper-6.1.0.tgz from _output_walm_charts saved to local.

2. Visit walm api https://server_host:31607/swagger, choose `POST /api/v1/release/{namespace}/withchart`.
3. Filled in the `namespace` && `release` field, upload zookeeper-6.1.0.tgz, in field `body`, with it empty or
ref [releaseRequest](docs/ref/releaseRequest-reference.md). Following is a body example:
   ```json
   {
     "name": "zk2",
     "configValues": {
        "appConfig": {
           "zookeeper": {
              "replicas": 3
           }
        }
     },
     "metaInfoParams": {},
     "dependencies": {},
     "releaseLabels": {}
   }
   ```
4. Filled in the `namespace` && `release` field, upload kafka-6.1.0.tgz, in field `body`, filled it with Following example,
   finally kafka will depend on existing zookeeper clusters.
   ```json
   {
     "name": "ka2",
     "dependencies": {
        "zookeeper": "zk2"
     }
   }
   ```
## Development
### Prerequisite
- Go 1.11+
### Getting the code
```
cd $GOPATH/src/WarpCloud
git clone https://github.com/WarpCloud/walm.git
cd walm
```
### Dependencies
The build uses dependencies in the vendor directory. 
Occasionally, you might need to update the dependencies.
```
glide up -v
```
### Building
```
make
```
### Testing
#### Unit Test
```
make test
```
#### E2E Test
##### Prerequisite
- K8s 1.9+
- Redis 2.8+
```
# config test/e2e_walm.yaml first
make e2e-test
```

# Usage
## Walm Server
### Prerequisite
- K8s 1.9+
- Redis 2.8+
### Start Server
```
# config walm.yaml first
export Pod_Namespace=<walmns> && export Pod_Name=<walmname> && walm serv --config walm.yaml
```
### Rest Api Swagger Ui
http://<server_host>:9001/swagger

## Walmcli
[walmcli使用说明](docs/walmcli.md)
```
walmctl --help
```

# Road Map
- Authentication & Authorization
- Release Status Real-Time Synchronization