// Package testsupport provides shared helpers for the database-backed
// integration suite: a real PostgreSQL instance (testcontainers), schema
// migration (goose) and per-test data cleanup. The implementation is compiled
// only under the `integration` build tag, so the default `go test ./...` run
// neither pulls in Docker nor the testcontainers dependency tree.
package testsupport
