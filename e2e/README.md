# Forwarder e2e tests

## Running the e2e tests with the test runner

1. Generate certificates `make -C certs certs`, this can be done once
1. Build the forwarder image `make -C ../ update-devel-image`, this needs to be done after each forwarder code change
2. Start the test runner `make run-e2e`
1. The test runner will run all the tests sequentially and output the results to the console
1. If one of the test fails, the procedure will stop, test output will be printed
1. Environment will not be pruned once the error occurred, remember to manually clean it up with `make down`
or use the containers to run single test and find bugs 

### Running specific test

Start the test runner `make run-e2e RUN=<test>` where `<test>` is a regex matching the test name.
It can be used with `SETUP` to run specific test in specific environment setup.

### Running tests for specific environment setup

Start the test runner `make run-e2e SETUP=<setup>` where `<setup>` is a regex matching the setup name.

The defaults setup naming scheme is `defaults-<httpbin-scheme>-<proxy-scheme>-<upstream-scheme>` ex. the `defaults-h2-http-https` setup will use 
* httpbin with http2 support, 
* proxy with http scheme,
* upstream with https scheme.

### Debugging

Start the test runner `make run-e2e SETUP=<setup> SETUP_ARGS="-debug"` where `<setup>` is the name of the setup you want to debug.
It would:
* enable debug logging in all containers,
* print test logs,
* preserve the environment after the test is finished.

After the test is finished:
* check the environment setup by looking at the `docker-compose.yml` file in the `e2e` directory,
* run `make dump-logs` to print all the logs to the console,
* use `docker-compose` or `docker` commands to inspect the running environment. 

The proxy service binds the following ports to the host:
- 3128 - the proxy port, use the proxy `curl -x <proxy-scheme>://localhost:3128 http://httpbin.org/get`, for https you may nedd to add `--proxy-insecure` flag
- 10000 - the API port, navigate to `http://localhost:10000` to see the API index page

### Testing for Go routine leaks

You can kill one of the containers with `make term` ex. `SRV=forwarder-e2e-httpbin-1 make term` to kill the httpbin container and see how the proxy behaves

### Using different forwarder image for testing

If one wants to use different forwarder image use `FORWARDER_VERSION` env variable ex. `FORWARDER_VERSION=1.0.0 make <target>`

## Benchmarking

1. Add benchmark to `bench_test.go` or use an existing one
1. Run the `make bench` function ex. `RUN=BenchmarkRespBody1k make bench` it will output the profile path directly to the console
1. Run pprof with the profile output ex. `go tool pprof -http=:8080 /path/to/profiles/cpu`  
