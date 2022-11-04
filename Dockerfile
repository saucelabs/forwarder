FROM ubuntu:22.10

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates curl && \
    apt-get clean && \
    apt-get autoremove && \
    rm -rf /var/lib/apt/lists/*

COPY forwarder /usr/bin/forwarder
ENTRYPOINT ["/usr/bin/forwarder"]
CMD ["proxy", "--api-address", ":10000"]

HEALTHCHECK --interval=1s --timeout=250ms --retries=10 CMD ["curl", "-s", "-S", "-f", "http://localhost:10000/readyz"]
