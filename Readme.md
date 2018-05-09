# Walm
The Warp application lifecycle manager,It use Helm as backend to create,delete,update,select instance, and serv restful api server.
# Usage
## start server
walm serv  [-a addr] [-p port]
## Build
### Builder
```
cd build/builder-docker && make
```
### Build
```
make swag && make build
```
## Test
### All
```
make test
```
### Unit test  [on going]
```
make unit-test
```
### E2E test  [on going]
```
make e2e-test
```
# Road Map
- [x] Helm RestFul Server
- [ ] Application Server
- [ ] Event Server