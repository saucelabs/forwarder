# forwarder ready

Usage: `forwarder ready [--api-address <host:port>] [flags]`

Readiness probe for the Forwarder.
This is equivalent to calling /readyz endpoint on the Forwarder API server.

## API server options

### `--api-address`

Environment variable: `FORWARDER_API_ADDRESS`

The API server address.

Default value: `localhost:10000`

