# Forwarder e2e tests

## Manually running the e2e tests

1. Build the forwarder image `make -C ../ update-devel-image`
1. Load test functions `source ./lib.sh`
1. Run the `run_test` function ex. `RUN=TestStatusCodes/400 run_test http http http`
1. Dump the logs if needed `make dump-logs` 