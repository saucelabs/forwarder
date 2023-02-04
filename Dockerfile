FROM gcr.io/distroless/static:nonroot

COPY LICENSE /licenses/
COPY LICENSE.3RD_PARTY /licenses/
COPY forwarder /usr/bin
ENTRYPOINT ["/usr/bin/forwarder"]
CMD ["proxy"]
