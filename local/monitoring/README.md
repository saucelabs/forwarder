# Monitoring

To enhance your development and testing environment, this package supports local Prometheus and Grafana instances.

## Starting and Stopping

Use the following commands to start and stop Prometheus and Grafana containers:
 - Start containers:
      ```bash
      make up
      ```
 - Stop containers:
   ```bash
   make down
   ```

## Prometheus Configuration

- The Prometheus instance is configured to scrape metrics from the following localhost ports:
    - `10000`

### Usage with End-to-End Tests

- Note that these ports are used by API server in end-to-end tests. This setup is designed to work seamlessly with end-to-end tests.

## Available dashboards

- Forwarder dashboard has been added to Grafana

## Accessing Monitoring Dashboards

- Prometheus: [http://localhost:9090](http://localhost:9090)
- Grafana: [http://localhost:3000](http://localhost:3000)

## Notes

- Ensure that ports `3000`, `9090`, `10000` are available on your system for Grafana, Prometheus, and Forwarder's API server, respectively.
- Customize Prometheus and Grafana configurations based on your requirements.
