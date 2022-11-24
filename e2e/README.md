# Forwarder e2e tests

## Manually running the e2e tests

1. Build the forwarder image `make -C ../ update-devel-image`
1. Load test functions `source ./lib.sh`
1. Run the `run_test` function ex. `RUN=TestStatusCodes/400 run_test http http http`
1. Dump the logs if needed `make dump-logs` 

Once the test is complete you may also run curl from the proxy container ex. `docker-compose exec proxy curl -vvv --insecure --proxy-insecure --proxy https://localhost:3128 https://httpbin/status/200`

## Test development

1. Run ./dev.sh to start the test environment with proxy and httpbin running on HTTP
1. Export the `DEV=1` environment variable to run the tests against the dev environment
