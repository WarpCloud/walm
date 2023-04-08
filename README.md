[1]: resource/Walm_Arch.jpg
[3]: resource/walm_logo.png

# Walm
![logo][3]

Walm is a micro service, based on Helm, that supports both Rest Api and Cli to manage the lifecycle of pod based applications in kubernetes cluster including those with dependencies.

Walm dynamically manages the dependencies of an application. An application can depend on the applications already existed, and the configurations of applications depending on would be injected automatically. Besides, once the configurations of applications depending on changes, the configurations would be injected again in real-time.

Walm supports more advanced Chart that use Jsonnet as template engine to render kubernetes objects. It is more suitable to orchestrate and deploy complex applications, such as Big Data applications.

Walm supports finely grained authentication and authorization, that would make one user only have relevant authorization under kubernetes namespace scope.

Walm uses a message system(Kafka) to synchronize the application's status in real-time. Once the application's status changes, Walm would produce an event to Kafka in real-time, and the Walm client would get the latest application status in real-time by consuming the Kafka event .

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

## [Get Started](docs/getting-started.md)
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
http://<server_host>:9001/swagger-ui/?url=/apidocs.json

## Walmcli
[walmcli使用说明](docs/walmcli.md)
```
walmctl --help
```

# Road Map
- Authentication & Authorization
- Release Status Real-Time Synchronization
