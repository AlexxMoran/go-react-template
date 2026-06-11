package user

import (
	"context"

	"github.com/yourorg/goapp/internal/platform/database/gen"
	"github.com/yourorg/goapp/internal/user/userapi"
)

// Module is the user module's public facade: a single concrete entry point that
// composes the read and write internals. Other modules depend on the slice of
// these methods they need by declaring their own narrow port interface (see
// auth.Users) and accept *Module at the composition root. The Module never
// performs authorization — that happens at the HTTP entrypoint.
type Module struct {
	repo    *repository
	queries *queries
}

// New builds the user module over the given database handle (a pool or a tx).
func New(db gen.DBTX) *Module {
	return &Module{repo: newRepository(db), queries: newQueries(db)}
}

func (m *Module) Create(ctx context.Context, p userapi.CreateParams) (userapi.User, error) {
	return m.repo.Create(ctx, p)
}

func (m *Module) UpdateProfile(ctx context.Context, id int64, firstName, lastName string) (userapi.User, error) {
	return m.repo.UpdateProfile(ctx, id, firstName, lastName)
}

func (m *Module) GetByID(ctx context.Context, id int64) (userapi.User, error) {
	return m.queries.GetByID(ctx, id)
}

func (m *Module) GetByEmail(ctx context.Context, email string) (userapi.User, error) {
	return m.queries.GetByEmail(ctx, email)
}

func (m *Module) List(ctx context.Context, skip, limit int) ([]userapi.User, error) {
	return m.queries.List(ctx, skip, limit)
}

func (m *Module) Count(ctx context.Context) (int64, error) {
	return m.queries.Count(ctx)
}
