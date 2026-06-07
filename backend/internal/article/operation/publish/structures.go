package publish

import "time"

// Snapshot is the set of facts the decision layer needs, captured by the gateway
// from the database. It is intentionally a plain value type with no ORM or
// driver types so the decision layer stays pure and testable.
type Snapshot struct {
	ArticleID  int64
	Status     string
	Title      string
	HasContent bool
}

// Decision is the plan produced by the pure decision layer and executed by the
// gateway. It carries every value the write needs, including the timestamp, so
// applying it is deterministic.
type Decision struct {
	ArticleID   int64
	NewStatus   string
	PublishedAt time.Time
}
