ARG BASE_IMAGE=gcr.io/distroless/static:nonroot

FROM ${BASE_IMAGE}

COPY LICENSE /licenses/
COPY LICENSE.3RD_PARTY /licenses/
COPY forwarder /usr/bin
ENTRYPOINT ["/usr/bin/forwarder"]
CMD ["run"]

ENV FORWARDER_API_ADDRESS="localhost:10000"
HEALTHCHECK --interval=1s --timeout=3s --retries=10 CMD ["/usr/bin/forwarder", "ready"]
