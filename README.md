# go-react-template

A full-stack starter template: a **Go backend** (layered, operation-oriented) and
a **React frontend** (FSD-lite, MobX + MUI), wired together and runnable with a
single `docker compose up`.

```
go-react-template/
├── backend/    Go API — chi · pgx/sqlc · JWT auth · policy/RBAC authz   (see backend/README.md)
├── frontend/   React 19 + TypeScript — Vite · MobX · MUI · axios        (see frontend/README.md)
└── docker-compose.yml   db (Postgres) + migrate (goose) + api + frontend
```

- **Backend**: see [backend/README.md](backend/README.md) for the architecture
  (operation pattern, policies, JWT) and how to add a domain.
- **Frontend**: a small authenticated shell — login / register / protected users
  directory — talking to the Go API.

---

## Quick start

```bash
make up
# first run creates backend/.env from the template, builds images,
# runs migrations, then starts everything.
```

| Service | URL |
|---|---|
| Frontend (Vite) | http://localhost:4200 |
| API | http://localhost:3000 |
| API health | http://localhost:3000/health |
| PostgreSQL | localhost:5433 |

Try it: open the frontend, **Register**, then **Login** — you land on the
protected users page that lists accounts from the Go backend.

Stop / reset:

```bash
make down       # stop
make clean      # stop + drop the database volume
make logs       # tail logs
```

Run backend tooling through the root Makefile:

```bash
make backend-test      # go test ./... (incl. architecture guardrails)
make backend-sqlc      # regenerate type-safe DB code
make backend-migrate-up
```

---

## How the two halves connect

- The Vite dev server proxies `/api/*` to the Go backend (`VITE_API_PROXY_TARGET`,
  set to `http://api:3000` in compose). Because the browser only ever talks to the
  frontend origin, there is **no CORS** and the refresh-token httpOnly cookie works
  out of the box.
- **Auth**: `POST /api/auth/login` returns an access token (kept in memory by MobX)
  and sets a refresh cookie. Axios attaches the bearer token; on a `401` with
  `message_key: "invalid_token"` it transparently calls `/api/auth/refresh`, rotates
  the token, and retries the original request.
- **Error contract**: the backend returns `{ "error": { message, message_key, fields } }`;
  the frontend surfaces `message` via snackbars and uses `message_key` for control flow.
- **Permissions**: list/detail responses include a `permissions` map computed from the
  backend policy, so the UI can show/hide actions without re-implementing the rules.

---

## Local development without Docker

Backend (needs Go 1.26+ and a Postgres):

```bash
cd backend
cp .env.example .env
docker compose up -d db          # or your own Postgres
DB_HOST=localhost make migrate-up
make run                          # API on :3000
```

Frontend (needs Node 22+):

```bash
cd frontend
npm install
npm run dev                       # Vite on :4200, proxies /api to localhost:3000
```

---

## Provenance

The backend is a Go port of a FastAPI/SQLAlchemy architecture; the frontend is
adapted from a React shell originally built against that FastAPI backend. The
adaptation aligned the API client to the Go contract (JSON login, `/auth/me`,
`/auth/refresh`, the `{error:{...}}` envelope) and dropped the email-verification
flow (the backend keeps an `is_verified` flag as an extension point).

## License

MIT.
