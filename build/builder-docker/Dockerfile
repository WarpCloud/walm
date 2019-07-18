FROM gcr.io/google_containers/kube-cross:v1.12.4-1 AS build-env
MAINTAINER TOS <tos@transwarp.io>

ADD . /go/src/WarpCloud/walm
RUN cd /go/src/WarpCloud/walm && make

FROM alpine
MAINTAINER TOS <tos@transwarp.io>

COPY --from=build-env /go/src/WarpCloud/walm/swagger-ui /swagger-ui
COPY --from=build-env /go/src/WarpCloud/walm/_output/walm /usr/bin/
COPY --from=build-env /go/src/WarpCloud/walm/test/ksonnet-lib /opt/ksonnet-lib
