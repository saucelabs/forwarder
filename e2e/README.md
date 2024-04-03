# Forwarder e2e tests

## Prerequisites

### Build the forwarder image

You can either use a ready image from the Docker Hub or build it yourself.
If using the Docker Hub image simply export FORWARDER_VERSION env variable with the version you want to use.

To build the image yourself run `make -C ../ update-devel-image`, this needs to be done after each forwarder code change.
You may want to export the `FORWARDER_VERSION=devel` env variable to the version you just built.
This will allow to use the compose command directly when debugging the tests.

### Generate certificates

Run `make -C certs certs` to generate the certificates needed for the tests.

## Running the tests

The simplest way to run the tests is to use the `make` command.
It will run all required steps and start the tests.

### Rebuilding the test image

If you want to rebuild the test image run `make update-test-image`, 
or `make update-test-image run-e2e` to rebuild the image and run the tests.

### Running specific test

Tests are grouped to setups and test cases.
Setup represents a specific environment configuration and list of test cases to run.

Setup naming follows a scheme, the most common setups are `defaults`.
The defaults setup naming scheme is `defaults-<httpbin-scheme>-<proxy-scheme>-<upstream-scheme>`, ex. `defaults-h2-http-https` setup will use
* httpbin with http2 support,
* proxy with http scheme,
* upstream with https scheme.

To run specific setups use `make run-e2e SETUP=<setup>` where `<setup>` is a regex matching the setup name.
To run specific test cases use `make run-e2e RUN=<test>` where `<test>` is a regex matching the test name.
Note that running specific test cases may require specific setup to be run.

## Debugging

In debug mode only single setup is run, and the environment is preserved after the test is finished.
Additionally, debug mode enables debug logging in all containers, and port forwarding on the proxy container.
The setup `compose.yaml` file is stored in the `e2e` directory.
The default debug compose project name is `forwarder-e2e`.

To run the tests in debug mode use `make debug SETUP=<setup> RUN=<test>` where `<setup>` is the name of the setup you want to debug and `<test>` is the name of the test you want to debug.
You may export COMPOSE_PROJECT_NAME env variable to set the project name to something else.
Once the test is finished you can access the environment with the provided project name.

For example to get the logs from the test run `docker compose -p forwarder-e2e -f compose.yaml logs`.

The following ports are exposed on localhost:
- 3128 - the proxy port, use the proxy `curl -x <proxy-scheme>://localhost:3128 http://httpbin.org/get`, for https you may need to add `--proxy-insecure` flag
- 10000 - the API port, navigate to `http://localhost:10000` to see the API index page

You may modify the `e2e/compose.yaml` file to expose additional ports or change the environment configuration.
To do that run `make down` to stop the environment, modify the file and run `make up` to start the environment again.

## Prometheus metrics

For Prometheus and Grafana setup, go to [local/monitoring](../local/monitoring) and run `make up`.
It will start Prometheus and Grafana containers and configure Prometheus to scrape the proxy metrics. 
