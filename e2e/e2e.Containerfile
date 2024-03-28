FROM alpine

RUN apk add --no-cache bash bash-completion ca-certificates curl jq

COPY certs/ca.crt /etc/forwarder/certs/ca.crt
COPY e2e.test /usr/bin/e2e.test

ENTRYPOINT ["/usr/bin/e2e.test"]
