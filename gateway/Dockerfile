FROM gportal/golang:latest

ARG RUNTIME_VERSION
ARG COMMIT_SHA1

# Import application source
COPY ./ /opt/app-root/src

# Change working directory
WORKDIR /opt/app-root/src

# Create /etc/gportal directories
RUN mkdir -p /etc/gportal/

# Build binary for Latency Service
RUN go build -v -o "${APP_ROOT}/gateway" cmd/main.go

# Finally delete application source
RUN rm -rf /opt/app-root/src/

RUN /usr/bin/fix-permissions ${APP_ROOT} && \
    /usr/bin/fix-permissions /etc/gportal/

CMD ["/opt/app-root/gateway"]
