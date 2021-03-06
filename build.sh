#!/bin/bash -x

# Get information
VERSION=`cat ./VERSION`
REVISION=`git rev-parse --abbrev-ref HEAD`
BRANCH=`git rev-parse --short HEAD`
DATE=`date -u +"%F-%T-%z%Z"`
OUT="${1:-./bin/awslogs-exporter}"
SRC="./cmd/awslogs-exporter/"

# Flags
#F_VER="-X github.com/houserater/awslogs-exporter/version.Version=${VERSION}"
#F_REV="-X github.com/houserater/awslogs-exporter/version.Revision=${REVISION}"
#F_BR="-X github.com/houserater/awslogs-exporter/version.Branch=${BRANCH}"
#F_DA="-X github.com/houserater/awslogs-exporter/version.BuildDate=${DATE}"
F_CMP="-w -linkmode external -extldflags '-static'"

go build -o ${OUT} --ldflags "${F_VER} ${F_REV} ${F_BR} ${F_DA} ${F_CMP}" ${SRC}
