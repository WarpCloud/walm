# Copyright 2017 The OpenPitrix Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

FROM 172.16.1.99/transwarp/walm-builder:1.0 as builder

WORKDIR /go/src/walm
COPY . .

RUN swag init walm && make build

FROM alpine:3.6
RUN apk add --update ca-certificates && update-ca-certificates
COPY --from=builder /go/bin/* /usr/local/bin/

ENTRYPOINT [ "entrypoint.sh" ] 