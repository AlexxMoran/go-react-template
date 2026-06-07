package publish

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/pkg/apperror"
)

// Scenario orchestrates the operation: open a transaction, load facts, decide,
// apply, commit. Authorization is the caller's responsibility and happens at the
// HTTP entrypoint before Run is invoked.
type Scenario struct {
	pool *pgxpool.Pool
}

func NewScenario(pool *pgxpool.Pool) *Scenario {
	return &Scenario{pool: pool}
}

// Run executes the publish operation and returns the updated row.
func (s *Scenario) Run(ctx context.Context, c Contract) (gen.Article, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return gen.Article{}, apperror.Internal(err)
	}
	defer tx.Rollback(ctx)

	gateway := NewGateway(tx)

	snapshot, err := gateway.Load(ctx, c)
	if err != nil {
		return gen.Article{}, err
	}

	decision, err := Decisions{}.Make(snapshot, time.Now())
	if err != nil {
		return gen.Article{}, err
	}

	result, err := gateway.Apply(ctx, decision)
	if err != nil {
		return gen.Article{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return gen.Article{}, apperror.Internal(err)
	}
	return result, nil
}
