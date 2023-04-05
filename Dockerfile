ARG GOLANG_VERSION=1.20

FROM docker.io/golang:${GOLANG_VERSION} as build

WORKDIR /azure-exporter

COPY go.* .
COPY main.go .
COPY collector .
COPY azure .

ARG VERSION=""
ARG COMMIT=""

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags "-X main.OSVersion=${VERSION} -X main.GitCommit=${COMMIT}" \
    -a -installsuffix cgo \
    -o /go/bin/azure-exporter \
    ./main.go

FROM gcr.io/distroless/static-debian11:latest

LABEL org.opencontainers.image.description "Prometheus Exporter for Azure"
LABEL org.opencontainers.image.source https://github.com/DazWilkin/azure-exporter

COPY --from=build /go/bin/azure-exporter /

EXPOSE 9999

ENTRYPOINT ["/azure-exporter"]