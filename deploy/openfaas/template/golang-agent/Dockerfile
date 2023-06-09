FROM --platform=${TARGETPLATFORM:-linux/amd64} ghcr.io/openfaas/of-watchdog:0.9.8 as watchdog
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.19-alpine as build

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

ENV CGO_ENABLED=0

RUN mkdir -p /go/src/handler
WORKDIR /go/src/handler
COPY . .

ARG GO111MODULE="on"
ARG GOFLAGS=""
ARG DEBUG=0

WORKDIR /go/src/handler/function

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go get 

WORKDIR /go/src/handler
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go get 

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build --ldflags "-s -w" -a -installsuffix cgo -o handler .

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3.17
# Add non root user and certs
RUN addgroup -S app && adduser -S -g app app
# Split instructions so that buildkit can run & cache
# the previous command ahead of time.
RUN mkdir -p /home/app \
    && chown app /home/app

WORKDIR /home/app

COPY --from=build --chown=app /go/src/handler/handler    .
COPY --from=build --chown=app /usr/bin/fwatchdog         .
COPY --from=build --chown=app /go/src/handler/function/  .

USER app

ENV fprocess="./handler"
ENV mode="http"
ENV upstream_url="http://127.0.0.1:8082"
ENV prefix_logs="false"

CMD ["./fwatchdog"]
