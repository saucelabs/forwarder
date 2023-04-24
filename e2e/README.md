# Forwarder e2e tests

## Running the e2e tests with the test runner

1. Build the forwarder image `make -C ../ update-devel-image`
1. Start the test runner `make run-e2e`
1. The test runner will run all the tests sequentially and output the results to the console
1. If one of the test fails, the procedure will stop, test output will be printed
1. Environment will not be pruned once the error occurred, remember to manually clean it up with `make down`
or use the containers to run single test and find bugs 

## Debugging / Manually running the e2e tests

1. Build the forwarder image `make -C ../ update-devel-image`
1. Provide `docker-compose.yaml` or use the one created by the test runner
1. Start the environment with `make up`
1. Run specific test with `RUN=<test> make test`
1. Dump containers logs if needed `make dump-logs` 
1. Stop the environment with `make down`

Once the test is complete you may also run curl from the proxy container ex. `docker-compose exec proxy curl -vvv --insecure --proxy-insecure --proxy https://localhost:3128 https://httpbin:8080/status/200`

You can kill one of the containers with `make term` ex. `SRV=forwarder-e2e-httpbin-1 make term` to kill the httpbin container and see how the proxy behaves

If one wants to use different forwarder image use `FORWARDER_VERSION` env variable ex. `FORWARDER_VERSION=1.0.0 make <target>`

## Benchmarking

1. Add benchmark to `bench_test.go` or use an existing one
1. Run the `make bench` function ex. `RUN=BenchmarkRespBody1k make bench` it will output the profile path directly to the console
1. Run pprof with the profile output ex. `go tool pprof -http=:8080 /path/to/profiles/cpu`  
