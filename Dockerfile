ARG GOLANG_VERSION=1.24.3

ARG COMMIT
ARG VERSION

ARG TARGETOS
ARG TARGETARCH

FROM --platform=${TARGETARCH} docker.io/golang:${GOLANG_VERSION} AS build

WORKDIR /azure-exporter

COPY go.* .
COPY main.go .
COPY collector collector
COPY azure azure

ARG TARGETOS
ARG TARGETARCH

ARG VERSION
ARG COMMIT

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
    -ldflags "-X main.OSVersion=${VERSION} -X main.GitCommit=${COMMIT}" \
    -a -installsuffix cgo \
    -o /go/bin/azure-exporter \
    ./main.go

FROM --platform=${TARGETARCH} gcr.io/distroless/static-debian11:latest

LABEL org.opencontainers.image.description="Prometheus Exporter for Azure"
LABEL org.opencontainers.image.source="https://github.com/DazWilkin/azure-exporter"

COPY --from=build /go/bin/azure-exporter /

EXPOSE 9999

ENTRYPOINT ["/azure-exporter"]
