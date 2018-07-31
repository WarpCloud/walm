FROM 172.16.1.99/transwarp/walm-builder:1.0 as builder

WORKDIR /go/src/walm
COPY . .

RUN swag init -g router/routers.go && make install

# kubectl and helm will be placed and build base image
FROM 172.16.1.99/gold/helm:tos18-latest
#RUN apk add --update ca-certificates && update-ca-certificates
RUN mkdir -p /etc/walm/conf
COPY --from=builder /go/bin/* /usr/local/bin/
COPY --from=builder /go/src/walm/pkg/setting/conf/* /etc/walm/conf/
ENV WALM_HOME=/root/.walm

CMD [ "walm","serv" ] 
#ENTRYPOINT [ "walm","serv" ]