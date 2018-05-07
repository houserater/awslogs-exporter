FROM houserater/awslogs-exporter_base:latest
USER root
RUN apk add --no-cache g++

USER awslogs-exporter

RUN go get github.com/golang/dep/cmd/dep
RUN go get github.com/golang/mock/mockgen
