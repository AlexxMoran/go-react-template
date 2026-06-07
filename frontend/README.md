# Frontend

Reusable React frontend, wired to the Go backend in this repo (`../backend`).

## Included

- React 19 + TypeScript + Vite (path aliases: `@app`, `@pages`, `@shared`)
- Material UI theme setup
- MobX root service container (`RootService`)
- Axios API layer: bearer-token, refresh-token rotation, and error interceptors
  matched to the Go `{ "error": { message, message_key } }` envelope
- Auth pages (login / register) and a `withAuth` protected-route helper
- Protected users directory page + users service (calls `GET /api/v1/users`)
- i18n bootstrap (en / ru)

## API contract

The Vite dev server proxies `/api` to the backend (`VITE_API_PROXY_TARGET`,
default `http://localhost:3000`). Endpoints used:

| Call | Endpoint |
|---|---|
| login | `POST /api/auth/login` `{ email, password }` → `{ access_token }` |
| register | `POST /api/auth/register` `{ email, password, first_name? }` |
| me | `GET /api/auth/me` |
| refresh | `POST /api/auth/refresh` (httpOnly cookie) |
| logout | `POST /api/auth/logout` |
| users | `GET /api/v1/users?skip=&limit=` |

## Quick Start

```bash
npm install
npm run dev      # http://localhost:4200, proxies /api to http://localhost:3000
```

## Useful Commands

```bash
npm run dev
npm run lint
npm run typecheck
npm run check
npm run build
```
