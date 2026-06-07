// Package publish implements the "publish an article" business operation. It is
// the Go translation of the Python operations/<verb_object>/ pattern:
//
//	Contract -> Scenario -> Gateway.Load -> Snapshot -> Decisions.Make ->
//	Decision -> Gateway.Apply -> Result
//
// Files:
//   - contract.go    input definition (no framework, no DB)
//   - structures.go  Snapshot (facts) and Decision (plan), plain values
//   - decisions.go   pure business rules — no I/O, unit-testable without a DB
//   - gateway.go     all DB reads/writes; translates rows <-> operation data
//   - scenario.go    orchestration inside a single transaction
package publish

// Contract is the input to the publish operation.
type Contract struct {
	ArticleID int64
}
