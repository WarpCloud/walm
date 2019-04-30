[1]: resource/Walm_Arch.jpg
[2]: resource/Walm_Server_Features.jpg

# Walm
The Warp application lifecycle manager,using Helm as backend to create,delete,update,get application, is composed of Walm Server and Walmctl.
Walm Server serves restful api server. Walmctl is cli for user.

## Architecture
![arch][1]

## Features
![feature][2]

## Build
```
make
```
## Test
### Unit Test
```
make test
```
### E2E Test
```
make e2e-test
```

# Walm Server
## Usage
### Start Server
export Pod_Namespace=<walmns> && export Pod_Name=<walmname> && walm serv --config walm.yaml

### Rest Api Swagger Ui
http://localhost:9001/swagger


# Walmcli
## Usage

# Road Map
- Authentication & Authorization