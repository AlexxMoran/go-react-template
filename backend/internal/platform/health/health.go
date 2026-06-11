// Package health provides framework-agnostic liveness/readiness checks. A
// Checker holds named dependency probes (e.g. a database ping); the HTTP layer
// renders the result. Liveness asks "is the process running?" (always yes if it
// can answer); readiness asks "can it serve traffic?" (all probes pass).
package health

import "context"

// Check reports the health of one dependency. A nil error means healthy.
type Check func(ctx context.Context) error

// Checker aggregates named dependency checks.
type Checker struct {
	checks map[string]Check
}

// New returns an empty Checker. Register dependencies with Register.
func New() *Checker {
	return &Checker{checks: make(map[string]Check)}
}

// Register adds a named check (e.g. "database"). Registering the same name twice
// replaces the previous check.
func (c *Checker) Register(name string, check Check) {
	c.checks[name] = check
}

// Run executes every registered check and reports overall readiness along with a
// per-dependency status map ("ok" or the error text). With no checks registered
// it reports ready.
func (c *Checker) Run(ctx context.Context) (ready bool, statuses map[string]string) {
	ready = true
	statuses = make(map[string]string, len(c.checks))
	for name, check := range c.checks {
		if err := check(ctx); err != nil {
			ready = false
			statuses[name] = err.Error()
			continue
		}
		statuses[name] = "ok"
	}
	return ready, statuses
}
