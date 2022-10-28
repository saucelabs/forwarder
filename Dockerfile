FROM ubuntu:22.10

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    apt-get clean && \
    apt-get autoremove && \
    rm -rf /var/lib/apt/lists/*

COPY forwarder /usr/bin/forwarder
ENTRYPOINT ["/usr/bin/forwarder"]
CMD ["proxy", "--api-address", ":10000"]
