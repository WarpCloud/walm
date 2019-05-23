[1]: resource/Walm_Arch.jpg
[3]: resource/walm_logo.png

# Walm
![logo][3]

The Warp application lifecycle manager,using Helm as backend to create,delete,update,get application, is composed of Walm Server and Walmctl.
Walm Server serves restful api server. Walmctl is cli for user.

## Architecture
![arch][1]

## Advantage
- Walm supports rest api to manage the lifecycle of applications.
- Walm supports the orchestration and deployment of complex applications.
- Walm supports the dynamic dependencies management.
- Walm supports the real-time synchronization of the application's status.
- Walm supports the finely grained authentication and authorization.
- Walm supports to retrieve the more detailed specification and status of applications.

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
# config walm.yaml first
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
```
walmctl --help
```

# Road Map
## Release 0.1 
- Authentication & Authorization
- Release Status Real-Time Synchronization
- Document