FROM 172.16.1.99/transwarp/walm-builder:1.0 as builder

WORKDIR /go/src/walm
COPY . .

RUN swag init -g router/routers.go && make build

# kubectl and helm will be placed and build base image
# base image  above
FROM alpine:3.6
RUN apk add --update ca-certificates && update-ca-certificates
COPY --from=builder /go/bin/* /usr/local/bin/
ENV WALM_HOME=/root/.walm
COPY conf/app.ini /root/.walm/app.ini

ENTRYPOINT [ "walm","serv" ] 