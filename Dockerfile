FROM golang:1.9-alpine


RUN apk --update add musl-dev gcc tar git bash wget && rm -rf /var/cache/apk/*

# Create user
ARG uid=1000
ARG gid=1000
RUN addgroup -g $gid awslogs-exporter
RUN adduser -D -u $uid -G awslogs-exporter awslogs-exporter

RUN mkdir -p /go/src/github.com/houserater/awslogs-exporter/
RUN chown -R awslogs-exporter:awslogs-exporter /go

WORKDIR /go/src/github.com/houserater/awslogs-exporter/

USER awslogs-exporter
