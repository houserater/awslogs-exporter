FROM houserater/awslogs-exporter_base:latest

USER root

# Prepare
WORKDIR /go/src/github.com/houserater/awslogs-exporter/
RUN mkdir -p /bin
COPY . ./

# Build
RUN ./build.sh /bin/awslogs-exporter
RUN chmod 755 /bin/awslogs-exporter

# Clean up
WORKDIR /
RUN rm -rf /go/src/*


EXPOSE 9223

ENTRYPOINT [ "/bin/awslogs-exporter" ]
CMD        [ "--help"]
