# Running in Docker

Forwarder is available as a Docker image from [Docker Hub](https://hub.docker.com/r/saucelabs/forwarder).

To run Forwarder in a container:

```bash
docker run --rm -it -p 3128:3128 saucelabs/forwarder
```

It's best to configure Forwarder in a container with environment variables.

```bash
docker run --rm -it -p 3128:3128 \
    -e FORWARDER_ADDRESS=:3128 \
    -e FORWARDER_PROXY=http://upstream:8081 \
    saucelabs/forwarder
```

For help run:

```bash
docker run --rm saucelabs/forwarder help
```
