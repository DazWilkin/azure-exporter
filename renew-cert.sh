#!/usr/bin/env bash

set -x

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

trap read debug

NAME="azure-exporter"
TODAY=$(date +%y%m%d)

mkdir ${PWD}/backups/${TODAY}

# key is constant between renewals: cp don't mv
# Back it up anyway for additional security
cp ${PWD}/azure-exporter.key ${PWD}/backups/${TODAY}
# Don't need crt+key as it can be regenerated (see below)
mv ${PWD}/azure-exporter.csr ${PWD}/backups/${TODAY}
mv ${PWD}/azure-exporter.crt ${PWD}/backups/${TODAY}

# Create CSR using private key
openssl req \
-new \
-key ${NAME}.key \
-out ${NAME}.csr

# Create new certificate using CSR and key
openssl x509 \
-req \
-days 365 \
-in ${NAME}.csr \
-signkey ${NAME}.key \
-out ${NAME}.crt \

# Cheated and uploaded crt using Azure Portal
