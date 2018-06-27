# Copyright 2017 The OpenPitrix Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

FROM golang:1.10.1-alpine3.7 as builder

RUN apk add --no-cache git curl openssl make gcc

RUN go get -u -v github.com/Masterminds/glide
RUN go get -u -v github.com/swaggo/swag
RUN cd /go/src/github.com/swaggo/swag && /usr/bin/make deps && /usr/bin/make

# swagger-0.13.0
# RUN go get github.com/go-swagger/go-swagger/cmd/swagger
RUN mkdir -p /swagger && cd /swagger \
	&& wget https://github.com/go-swagger/go-swagger/releases/download/0.13.0/swagger_linux_amd64 \
	&& chmod +x swagger_linux_amd64 && mv swagger_linux_amd64 /go/bin/swagger


FROM golang:1.10.1-alpine3.7

RUN apk add --no-cache git  make curl openssl jq gcc

COPY --from=builder /go/bin /go/bin
