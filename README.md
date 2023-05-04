# Prometheus Exporter for Azure

[![build](https://github.com/DazWilkin/azure-exporter/actions/workflows/build.yml/badge.svg)](https://github.com/DazWilkin/azure-exporter/actions/workflows/build.yml)

## Installation

The application uses [`DefaultAzureCredential`](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity@v1.2.2#readme-authenticate-with-defaultazurecredential) to authenticate using the developer's `az` identity.

## [Sigstore](https://www.sigstore.dev)

`azure-exporter` container images are being signed by Sigstore and may be verified:

```bash
cosign verify \
--key=./cosign.pub \
ghcr.io/dazwilkin/azure-exporter:6aaffdbab0969afe61907f21a856684f00b43290
```

## Go

Uses `azidentity.NewDefaultAzureCredential`, please `az login` to ensure credentials are available before running:

```bash
SUBSCRIPTION="..." # Azure Subscription ID

PORT="8080"

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
-keyout azure-exporter.key \
-out azure-exporter.crt \
-sha256 \
-days 365 \
-nodes \
-subj "/CN=azure-exporter"
```

In a subsequent step, the Azure CLI will set `AZURE_CLIENT_CERTIFICATE_PATH` to point to a file that contains **both** the key and cert:

```bash
cat azure-exporter.key >> azure-exporter.key+crt
cat azure-exporter.crt >> azure-exporter.key+crt
```

Then:

```bash
SUBSCRIPTION="..."
GROUP="..."

NAME="azure-exporter"

az ad sp create-for-rbac \
--name=${NAME} \
--role="Reader" \
--scopes="/subscriptions/${SUBSCRIPTION}/resourceGroups/${GROUP}" \
--cert=@${PWD}/${NAME}.crt
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
AZURE_CLIENT_CERTIFICATE_PATH="${PWD}/azure-exporter.key+crt"

PORT="8080"

podman run \
--interactive --tty --rm \
--name=azure-exporter \
--env=SUBSCRIPTION=${SUBSCRIPTION} \
--env=AZURE_CLIENT_ID=${AZURE_CLIENT_ID} \
--env=AZURE_TENANT_ID=${AZURE_TENANT_ID} \
--env=AZURE_CLIENT_CERTIFICATE_PATH=/secrets/azure-exporter.key+crt \
--volume=${AZURE_CLIENT_CERTIFICATE_PATH}:/secrets/azure-exporter.key+crt \
--publish=${PORT}:${PORT}/tcp \
ghcr.io/dazwilkin/azure-exporter:6aaffdbab0969afe61907f21a856684f00b43290 \
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
groups:
- name: azure_exporter
  rules:
  - alert: azure_container_apps_running
    expr: min_over_time(azure_container_apps_total{}[15m]) > 0
    for: 6h
    labels:
      severity: page
    annotations:
      summary: "Azure Container Apps ({{ $value }}) running (resource group: {{ $labels.resourcegroup }})"
```

<hr/>
<br/>
<a href="https://www.buymeacoffee.com/dazwilkin" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/default-orange.png" alt="Buy Me A Coffee" height="41" width="174"></a>
