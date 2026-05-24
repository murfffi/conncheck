# conncheck - the missing `net.Conn.IsOpen` in Go

conncheck implements robust, fast, and reusable network connection liveness check in Go.
Essentially, `conncheck.Do` implements the missing `net.Conn.IsOpen()` method that tells you
if the connection has been closed on the client, was interrupted by network infrastructure, 
or the server asked to close it e.g., with an RST or FIN packet in case of TCP.

`conncheck` is useful in any network application with long-lived, sometimes idle, connections
like SQL connection pools, Apache Thrift clients, etc.

[![Go Reference](https://pkg.go.dev/badge/github.com/murfffi/conncheck.svg)](https://pkg.go.dev/github.com/murfffi/conncheck)
[![Go Report Card](https://goreportcard.com/badge/github.com/murfffi/conncheck)](https://goreportcard.com/report/github.com/murfffi/conncheck)
[![Tests](https://github.com/murfffi/conncheck/actions/workflows/ci.yml/badge.svg)](https://coveralls.io/github/murfffi/conncheck)

## Quickstart

```go
import "github.com/murfffi/conncheck"

func IsBroken(conn net.Conn) bool {
    return conncheck.Do(conn) == conncheck.StatusNotOpen
}
```

## How it works

`conncheck` builds on network liveness checks implemented in some Go SQL drivers, replacing
expensive server "ping" calls. While server pings more accurately determine if a connection is live,
the ping takes the time of a full round-trip and consumes precious server resources. Instead,
`conncheck` peeks without blocking from the connection at OS level. If the OS kernel is aware
of the connection being closed or interrupted, by either peer, it will reject the peek.

The Github team first documented the general approach in 
[a blog post](https://github.blog/engineering/three-bugs-in-the-go-mysql-driver/)
The approach was later borrowed by other libraries that use  long-lived connections 
like [Apache Thrift](https://github.com/apache/thrift/pull/2153).

Unlike existing solutions, `conncheck` is reusable and supports both Unix systems and Windows.
PRs will be opened, using this code, for the OSS projects with existing limited solutions.

## Copyright and acknowledgements

This library is licensed under the Apache 2 license. All code was written from scratch. The Unix code
used the Apache Thrift implementation as a reference, while the Windows code is completely novel. 
