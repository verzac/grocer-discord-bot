# GroceryBot (GroBot)

A Discord bot that maintains grocery lists per-guild, using Go, discordgo, GORM, SQLite, and Echo v4. The `/docs` subfolder is an independent Docusaurus site.

## Cursor Cloud specific instructions

### Go bot (main service)

- **Run**: `go run main.go` or `make air` (hot-reload via `air -c .air.toml`; requires `air` on PATH — add `$(go env GOPATH)/bin` to PATH).
- **Build**: `go build -o tmp/main main.go` (CGO required for SQLite — gcc must be installed).
- **Lint/vet**: `go vet ./...`
- **Tests**: Only E2E tests exist (`make e2e`); they require two Discord bot tokens and a real Discord server. There are no unit tests.
- The bot **panics on startup** if `GROCER_BOT_TOKEN` is not set. A valid Discord bot token is required for the bot to stay running.
- SQLite DB is embedded — no external database process needed. Migrations run automatically from `db/changelog/` on startup.
- The REST API starts on `:8080` (in a goroutine) before Discord connects. With an invalid token, the process exits almost immediately after API starts (~250ms window).
- Environment variables are loaded from `.env` via `godotenv`. See `README.md` for the full list.
- The `db/gorm.db` file is auto-created on first run; the `db/` directory must exist.

### Docs site (`/docs`)

- Uses **yarn** (lockfile: `yarn.lock`).
- **Dev server**: `cd docs && yarn start` (port 3000).
- **Build**: `cd docs && yarn build`.
- No tests for the docs site.

### E2E tests

- Require env vars: `GROCER_BOT_TOKEN`, `E2E_BOT_TOKEN`, `E2E_CHANNEL_ID`, `E2E_GROCER_BOT_ID`, `E2E_GUILD_ID`.
- Run with `make e2e` (bot must be running) or `make full_e2e` (builds, starts bot, runs tests, cleans up).

### PATH note

`air` installs to `$(go env GOPATH)/bin` (typically `/home/ubuntu/go/bin`). This directory must be on PATH for `make air` to work. Add it in your shell if not already present:
```
export PATH="$PATH:$(go env GOPATH)/bin"
```
