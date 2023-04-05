# Prometheus Exporter for Azure

## Installation

The application uses [`DefaultAzureCredential`](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity@v1.2.2#readme-authenticate-with-defaultazurecredential) to authenticate using the developer's `az` identity.

## [Sigstore](https://www.sigstore.dev)

`azure-exporter` container images are being signed by Sigstore and may be verified:

```bash
cosign verify \
--key=./cosign.pub \
ghcr.io/dazwilkin/azure-exporter:1234567890123456789012345678901234567890
```

## Standalone

```bash
PORT=9999
REPO="ghcr.io/dazwilkin/azure-exporter"

podman run \
--interactive --tty --rm \
--publish=${PORT}:${PORT}/tcp \
ghcr.io/dazwilkin/azure-exporter:1234567890123456789012345678901234567890 \
--endpoint=0.0.0.0:${PORT}
--metrics_path="/metrics"
```

## Metrics

|Name|Type|Description|
|----|----|-----------|


## Prometheus

