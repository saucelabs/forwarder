FROM gcr.io/distroless/static:nonroot

COPY forwarder /usr/bin/forwarder
ENTRYPOINT ["/usr/bin/forwarder"]
CMD ["proxy"]
