---
id: metrics
title: Metrics
---

# Prometheus Metrics

## forwarder run

### `forwarder_dialer_closed_total`

Number of closed connections

Labels:
  - host

### `forwarder_dialer_dialed_total`

Number of dialed connections

Labels:
  - host

### `forwarder_dialer_errors_total`

Number of dialer errors

Labels:
  - host

### `forwarder_errors_total`

Number of errors

Labels:
  - name

### `forwarder_http_request_duration_seconds`

The HTTP request latencies in seconds.

Labels:
  - code
  - method

### `forwarder_http_request_size_bytes`

The HTTP request sizes in bytes.

Labels:
  - code
  - method

### `forwarder_http_requests_in_flight`

Current number of HTTP requests being served.

Labels:
  - method

### `forwarder_http_requests_total`

Total number of HTTP requests processed.

Labels:
  - code
  - method

### `forwarder_http_response_size_bytes`

The HTTP response sizes in bytes.

Labels:
  - code
  - method

### `forwarder_listener_accepted_total`

Number of accepted connections

### `forwarder_listener_closed_total`

Number of closed connections

### `forwarder_listener_errors_total`

Number of listener errors when accepting connections

### `forwarder_listener_tls_errors_total`

Number of TLS handshake errors

### `forwarder_process_cpu_seconds_total`

Total user and system CPU time spent in seconds.

### `forwarder_process_max_fds`

Maximum number of open file descriptors.

### `forwarder_process_open_fds`

Number of open file descriptors.

### `forwarder_process_resident_memory_bytes`

Resident memory size in bytes.

### `forwarder_process_start_time_seconds`

Start time of the process since unix epoch in seconds.

### `forwarder_process_virtual_memory_bytes`

Virtual memory size in bytes.

### `forwarder_process_virtual_memory_max_bytes`

Maximum amount of virtual memory available in bytes.

### `forwarder_proxy_errors_total`

Number of proxy errors

Labels:
  - reason

### `forwarder_version`

Forwarder version, value is always 1

Labels:
  - commit
  - time
  - version

### `go_env_gomaxprocs`

Number of maximum goroutines that can be executed simultaneously

### `go_env_gomemlimit`

Memory limit for the process

### `go_gc_duration_seconds`

A summary of the pause duration of garbage collection cycles.

### `go_goroutines`

Number of goroutines that currently exist.

### `go_info`

Information about the Go environment.

Labels:
  - version

### `go_memstats_alloc_bytes`

Number of bytes allocated and still in use.

### `go_memstats_alloc_bytes_total`

Total number of bytes allocated, even if freed.

### `go_memstats_buck_hash_sys_bytes`

Number of bytes used by the profiling bucket hash table.

### `go_memstats_frees_total`

Total number of frees.

### `go_memstats_gc_sys_bytes`

Number of bytes used for garbage collection system metadata.

### `go_memstats_heap_alloc_bytes`

Number of heap bytes allocated and still in use.

### `go_memstats_heap_idle_bytes`

Number of heap bytes waiting to be used.

### `go_memstats_heap_inuse_bytes`

Number of heap bytes that are in use.

### `go_memstats_heap_objects`

Number of allocated objects.

### `go_memstats_heap_released_bytes`

Number of heap bytes released to OS.

### `go_memstats_heap_sys_bytes`

Number of heap bytes obtained from system.

### `go_memstats_last_gc_time_seconds`

Number of seconds since 1970 of last garbage collection.

### `go_memstats_lookups_total`

Total number of pointer lookups.

### `go_memstats_mallocs_total`

Total number of mallocs.

### `go_memstats_mcache_inuse_bytes`

Number of bytes in use by mcache structures.

### `go_memstats_mcache_sys_bytes`

Number of bytes used for mcache structures obtained from system.

### `go_memstats_mspan_inuse_bytes`

Number of bytes in use by mspan structures.

### `go_memstats_mspan_sys_bytes`

Number of bytes used for mspan structures obtained from system.

### `go_memstats_next_gc_bytes`

Number of heap bytes when next garbage collection will take place.

### `go_memstats_other_sys_bytes`

Number of bytes used for other system allocations.

### `go_memstats_stack_inuse_bytes`

Number of bytes in use by the stack allocator.

### `go_memstats_stack_sys_bytes`

Number of bytes obtained from system for stack allocator.

### `go_memstats_sys_bytes`

Number of bytes obtained from system.

### `go_threads`

Number of OS threads created.

