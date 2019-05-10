FROM 172.16.1.99/transwarp/gcr.io/google_containers/kube-cross:v1.12.4-1 AS build-env
MAINTAINER TOS <tos@transwarp.io>

ADD . /go/src/walm
RUN cd /go/src/walm && make

FROM 172.16.1.99/transwarp/alpine:transwarp-base
MAINTAINER TOS <tos@transwarp.io>

COPY --from=build-env /go/src/walm/swagger-ui /swagger-ui
COPY --from=build-env /go/src/walm/_output/walm /usr/bin/
COPY --from=build-env /go/src/walm/test/ksonnet-lib /opt/ksonnet-lib

