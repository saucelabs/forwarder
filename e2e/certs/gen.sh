#!/usr/bin/env bash

set -eu -o pipefail

# Common variables
CA_ORGANIZATION="Sauce Labs Inc."
CA_KEY="ca.key"
CA_CERT="ca.crt"
CA_SUBJECT="/C=US/O=${CA_ORGANIZATION}"

EC_CURVE="prime256v1"

# Generate CA key and self-signed certificate with SHA-256
openssl ecparam -genkey -name ${EC_CURVE} -out ${CA_KEY}
openssl req -new -x509 -sha256 -days 365 -nodes -key ${CA_KEY} -subj "${CA_SUBJECT}" -out ${CA_CERT} \
-extensions v3_ca -config <(cat /etc/ssl/openssl.cnf - << EOF
[v3_ca]
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid:always,issuer
basicConstraints=critical,CA:true
keyUsage=critical,keyCertSign,cRLSign
EOF
)

# Function to generate certificates for each host name
generate_certificate() {
    local HOST_NAME="$1"
    local KEY="${HOST_NAME}.key"
    local CSR="${HOST_NAME}.csr"
    local CERT="${HOST_NAME}.crt"
    local SUBJECT="/C=US/O=${CA_ORGANIZATION}/CN=${HOST_NAME}"

    # Generate host key and certificate signing request (CSR)
    openssl ecparam -genkey -name ${EC_CURVE} -out ${KEY}
    openssl req -new -key ${KEY} -subj "${SUBJECT}" -out ${CSR}

    # Sign the CSR with the CA to generate the host certificate
    openssl x509 -req -sha256 -days 365 -in ${CSR} -CA ${CA_CERT} -CAkey ${CA_KEY} -CAcreateserial -out ${CERT}\
    -extensions v3_req -extfile <(cat /etc/ssl/openssl.cnf - << EOF
[v3_req]
basicConstraints=critical,CA:FALSE
authorityKeyIdentifier=keyid,issuer
subjectAltName=@alt_names
keyUsage=digitalSignature,keyEncipherment
[ alt_names ]
DNS.1 = ${HOST_NAME}
DNS.2 = localhost
EOF
    )

    # Remove the CSR (not needed anymore)
    rm ${CSR}
}

# Generate certificates for each host name
generate_certificate "proxy"
generate_certificate "upstream-proxy"
generate_certificate "httpbin"

chmod 644 *.key *.crt
