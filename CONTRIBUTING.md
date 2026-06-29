# Contributing to OpenWard

Thanks for helping improve OpenWard. This project is intentionally small and focused: a Go reverse proxy at the root, plus a Svelte/Vite UI in `ui/`. Contributions are easiest to review when they stay aligned with that shape.

## Project Layout

- `cmd/` contains executable entry points.
- `internal/` contains the Go backend packages.
- `ui/` contains the Svelte frontend and its own package tooling.
- `benchmarks/` contains performance scripts.

## Setup

### Backend

```bash
go mod download
cp .env.example .env
```

The backend reads these environment variables from `.env` if present:

- `OPENWARD_ADMIN_EMAIL`
- `OPENWARD_ADMIN_PASSWORD`
- `OPENWARD_ADDR`
- `OPENWARD_ADMIN_ADDR`

Run the service with:

```bash
go run ./cmd
```

If you want a seeded local project for proxy testing:

```bash
go run ./cmd/seed/main.go
```

### Frontend

```bash
cd ui
npm install
npm run dev
```

## Validation

Before opening a pull request, run the checks that match the area you changed.

### Go backend

```bash
gofmt -w <files>
go test ./...
go run ./cmd
```

If your change touches database or proxy behavior, also verify a request through the proxy with the seeded project.

### UI

```bash
cd ui
npm run check
npm run build
```

## Code Style

- Keep Go code formatted with `gofmt`.
- Prefer small, focused changes over broad refactors.
- Keep packages narrow and use `internal/` for backend implementation details.
- Match the existing Svelte style in `ui/src/` and avoid introducing extra frameworks or routing libraries.
- Use descriptive names for rate limiting algorithms, project settings, and admin/auth flows.

## Tests and Benchmarks

There is not a large automated test suite yet, so correctness matters in review.

- Add tests when you introduce new backend behavior or fix a bug that can be reproduced.
- Keep benchmark updates in `benchmarks/` reproducible and explain what changed if results move.
- Do not commit generated benchmark output unless it is intentionally part of the change.

## Pull Request Checklist

- Describe what changed and why.
- Call out any config or migration impact.
- Include the commands you ran to verify the change.
- Mention follow-up work if the change is intentionally incomplete.

## Security Notes

- Do not commit real credentials or upstream API keys.
- Treat `.env` and local SQLite databases as machine-specific files.
- If you change auth, proxy forwarding, or rate limiting, include a careful explanation of the behavior change.

