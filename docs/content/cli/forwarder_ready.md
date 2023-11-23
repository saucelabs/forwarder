---
title: forwarder ready
weight: 104
---

# Forwarder Ready

Usage: `forwarder ready [--api-address <host:port>] [flags]`

Readiness probe for the Forwarder.
This is equivalent to calling /readyz endpoint on the Forwarder API server.

**Note:** You can also specify the options as YAML, JSON or TOML file using `--config-file` flag.
You can generate a config file by running `forwarder ready config-file` command.


## API server options

### `--api-address`

Environment variable: `FORWARDER_API_ADDRESS`

The API server address.

Default value: `localhost:10000`

