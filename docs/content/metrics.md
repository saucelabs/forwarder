---
id: metrics
title: Metrics
---

# Prometheus Metrics

## forwarder run

### `forwarder_dialer_cx_active`

Number of active connections

Labels:
  - host

### `forwarder_dialer_cx_total`

Number of dialed connections

Labels:
  - host

### `forwarder_dialer_errors_total`

Number of errors dialing connections

Labels:
  - host

### `forwarder_dialer_retries_total`

Number of dial retries

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

### `forwarder_http_requests_in_flight`

Current number of HTTP requests being served.

Labels:
  - method

### `forwarder_http_requests_total`

Total number of HTTP requests processed.

Labels:
  - code
  - method

### `forwarder_listener_cx_active`

Number of active connections

### `forwarder_listener_cx_total`

Number of accepted connections

### `forwarder_listener_errors_total`

Number of listener errors when accepting connections

### `forwarder_process_cpu_seconds_total`

Total user and system CPU time spent in seconds.

### `forwarder_process_max_fds`

Maximum number of open file descriptors.

### `forwarder_process_network_receive_bytes_total`

Number of bytes received by the process over the network.

### `forwarder_process_network_transmit_bytes_total`

Number of bytes sent by the process over the network.

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

A summary of the wall-time pause (stop-the-world) duration in garbage collection cycles.

### `go_gc_gogc_percent`

Heap size target percentage configured by the user, otherwise 100. This value is set by the GOGC environment variable, and the runtime/debug.SetGCPercent function. Sourced from /gc/gogc:percent

### `go_gc_gomemlimit_bytes`

Go runtime memory limit configured by the user, otherwise math.MaxInt64. This value is set by the GOMEMLIMIT environment variable, and the runtime/debug.SetMemoryLimit function. Sourced from /gc/gomemlimit:bytes

### `go_goroutines`

Number of goroutines that currently exist.

### `go_info`

Information about the Go environment.

Labels:
  - version

### `go_memstats_alloc_bytes`

Number of bytes allocated in heap and currently in use. Equals to /memory/classes/heap/objects:bytes.

### `go_memstats_alloc_bytes_total`

Total number of bytes allocated in heap until now, even if released already. Equals to /gc/heap/allocs:bytes.

### `go_memstats_buck_hash_sys_bytes`

Number of bytes used by the profiling bucket hash table. Equals to /memory/classes/profiling/buckets:bytes.

### `go_memstats_frees_total`

Total number of heap objects frees. Equals to /gc/heap/frees:objects + /gc/heap/tiny/allocs:objects.

### `go_memstats_gc_sys_bytes`

Number of bytes used for garbage collection system metadata. Equals to /memory/classes/metadata/other:bytes.

### `go_memstats_heap_alloc_bytes`

Number of heap bytes allocated and currently in use, same as go_memstats_alloc_bytes. Equals to /memory/classes/heap/objects:bytes.

### `go_memstats_heap_idle_bytes`

Number of heap bytes waiting to be used. Equals to /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.

### `go_memstats_heap_inuse_bytes`

Number of heap bytes that are in use. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes

### `go_memstats_heap_objects`

Number of currently allocated objects. Equals to /gc/heap/objects:objects.

### `go_memstats_heap_released_bytes`

Number of heap bytes released to OS. Equals to /memory/classes/heap/released:bytes.

### `go_memstats_heap_sys_bytes`

Number of heap bytes obtained from system. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes + /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.

### `go_memstats_last_gc_time_seconds`

Number of seconds since 1970 of last garbage collection.

### `go_memstats_mallocs_total`

Total number of heap objects allocated, both live and gc-ed. Semantically a counter version for go_memstats_heap_objects gauge. Equals to /gc/heap/allocs:objects + /gc/heap/tiny/allocs:objects.

### `go_memstats_mcache_inuse_bytes`

Number of bytes in use by mcache structures. Equals to /memory/classes/metadata/mcache/inuse:bytes.

### `go_memstats_mcache_sys_bytes`

Number of bytes used for mcache structures obtained from system. Equals to /memory/classes/metadata/mcache/inuse:bytes + /memory/classes/metadata/mcache/free:bytes.

### `go_memstats_mspan_inuse_bytes`

Number of bytes in use by mspan structures. Equals to /memory/classes/metadata/mspan/inuse:bytes.

### `go_memstats_mspan_sys_bytes`

Number of bytes used for mspan structures obtained from system. Equals to /memory/classes/metadata/mspan/inuse:bytes + /memory/classes/metadata/mspan/free:bytes.

### `go_memstats_next_gc_bytes`

Number of heap bytes when next garbage collection will take place. Equals to /gc/heap/goal:bytes.

### `go_memstats_other_sys_bytes`

Number of bytes used for other system allocations. Equals to /memory/classes/other:bytes.

### `go_memstats_stack_inuse_bytes`

Number of bytes obtained from system for stack allocator in non-CGO environments. Equals to /memory/classes/heap/stacks:bytes.

### `go_memstats_stack_sys_bytes`

Number of bytes obtained from system for stack allocator. Equals to /memory/classes/heap/stacks:bytes + /memory/classes/os-stacks:bytes.

### `go_memstats_sys_bytes`

Number of bytes obtained from system. Equals to /memory/classes/total:byte.

### `go_sched_gomaxprocs_threads`

The current runtime.GOMAXPROCS setting, or the number of operating system threads that can execute user-level Go code simultaneously. Sourced from /sched/gomaxprocs:threads

### `go_threads`

Number of OS threads created.

