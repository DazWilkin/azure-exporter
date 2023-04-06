# Prometheus Exporter for Azure

[![build](https://github.com/DazWilkin/azure-exporter/actions/workflows/build.yml/badge.svg)](https://github.com/DazWilkin/azure-exporter/actions/workflows/build.yml)

## Installation

The application uses [`DefaultAzureCredential`](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity@v1.2.2#readme-authenticate-with-defaultazurecredential) to authenticate using the developer's `az` identity.

## [Sigstore](https://www.sigstore.dev)

`azure-exporter` container images are being signed by Sigstore and may be verified:

```bash
cosign verify \
--key=./cosign.pub \
ghcr.io/dazwilkin/azure-exporter:2a7243d13c2e47bfab2cf30aacd92b4c1bf3d5b8
```

## Go

Uses `azidentity.NewDefaultAzureCredential`, please `az login` to ensure credentials are available before running:

```bash
SUBSCRIPTION="..." # Azure Subscription ID

PORT="9746"

go run github.com/DazWilkin/azure-exporter \
--endpoint="0.0.0.0:${PORT}" \
--path="/metrics"
```

**NOTE**
1. `go run .` works too
1. `--endpoint` defaults to `0.0.0.0:9476` and `--path` defaults to `/metrics` so both arguments are redundant

## Container

When running in a (Linux) container, the exporter is unable to obtain CLI (`az login`) credentials.

Please create a Service Principal and use its credentials:

First, you'll need a certificate and key:

```bash
openssl req \
-x509 \
-newkey rsa:4096 \
-keyout key.pem \
-out crt.pem \
-sha256 \
-days 365 \
-nodes \
-subj "/CN=azure-exporter"
```

This will generate `crt.pem` and `key.pem`.

In a subsequent step, the Azure CLI will set `AZURE_CLIENT_CERTIFICATE_PATH` to point to a file that contains **both** the key and cert:

```bash
cat key.pem >> key+crt.pem
cat crt.pem >> key+crt.pem
```
Then:
```bash
SUBSCRIPTION="..."
GROUP="..."
NAME="..."

az ad sp create-for-rbac \
--name=${NAME} \
--role="Reader" \
--scopes="/subscriptions/${SUBSCRIPTION}/resourceGroups/${GROUP}" \
--cert=@${PWD}/crt.pem
```
Yields:
```JSON
{
  "appId": "{AZURE_CLIENT_ID}",
  "displayName": "{NAME}",
  "password": null,
  "tenant": "{AZURE_TENANT_ID}"
}
```



Then, using the above-generated values for the environment variables shown below, you can run the container:

```bash
SUBSCRIPTION="..." # Azure Subscription ID

AZURE_CLIENT_ID="..." # Use values from Service Principal
AZURE_TENANT_ID="..."
AZURE_CLIENT_CERTIFICATE_PATH=".."

PORT="9746"

podman run \
--interactive --tty --rm \
--name=azure-exporter \
--env=SUBSCRIPTION=${SUBSCRIPTION} \
--env=AZURE_CLIENT_ID=${AZURE_CLIENT_ID} \
--env=AZURE_TENANT_ID=${AZURE_TENANT_ID} \
--env=AZURE_CLIENT_CERTIFICATE_PATH=/secrets/crt+key.pem \
--volume=${AZURE_CLIENT_CERTIFICATE_PATH}:/secrets/crt+key.pem \
--publish=${PORT}:${PORT}/tcp \
ghcr.io/dazwilkin/azure-exporter:2a7243d13c2e47bfab2cf30aacd92b4c1bf3d5b8 \
--endpoint=0.0.0.0:${PORT} \
--path="/metrics"
```

## Metrics

|Name|Type|Description|
|----|----|-----------|
|`azure_container_apps_total`|Gauge|Number of Azure Container Apps deployed|
|`azure_exporter_build_info`|Counter|Describes build info|
|`azure_exporter_start_time`|Gauge|The time (UNIX epoch) when the exporter started|
|`azure_resource_groups_total`|Gauge|Number of Azure Resource Groups|

## Prometheus

## AlertManager

For example:

```YAML

```