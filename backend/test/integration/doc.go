// Package integration holds the database-backed behavioral test suite. The tests
// drive the modules through their public seams (the application facades and the
// auth service) against a real PostgreSQL started with testcontainers, asserting
// behavior that unit tests cannot — real SQL, transactions and constraints.
//
// They are compiled only under the `integration` build tag; run them with:
//
//	go test -tags=integration ./test/...   (or: make test-integration)
package integration
