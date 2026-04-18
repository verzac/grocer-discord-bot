# SPEC-001 Progress Tracker

> Last updated: 2026-04-18

Reference spec: `.cursor/specs/SPEC-001-auth-rework-for-apps/spec.md`

---

## Decision: OAuth2 Flow Architecture (Option A — Backend Callback)

The Discord OAuth2 redirect URI points to the **GroceryBot backend** (`api.grocerybot.net/auth/discord/callback`), not the app directly. After exchanging the code for tokens and issuing a GroceryBot JWT, the backend redirects the browser to the app via a custom scheme (`grocerybot://auth/callback?access_token=...&refresh_token=...`). This keeps the `DISCORD_CLIENT_SECRET` server-side and simplifies redirect URI registration with Discord.

---

## Thin slice: Areas 1, 4, 5 (historical)

Implemented first: Discord OAuth2 login → GroceryBot JWT + refresh token; refresh endpoint. Discord tokens and PKCE were added in a follow-up (see below).

| Item | Location |
|------|----------|
| OAuth2 config + `APP_REDIRECT_URI` | `auth/oauth2.go` |
| Refresh token generate + SHA-256 hash | `auth/refresh_token.go` |
| `GET /auth/discord`, `GET /auth/discord/callback`, `POST /auth/refresh` | `handlers/api/routeauth/auth_handlers.go` |
| Register auth routes when Discord env complete | `handlers/api/routes.go` |
| Skip auth for `/auth/*` | `handlers/api/middleware/auth_middleware.go` |
| Minimal `user_sessions` migration + model + repo | `db/changelog/000011_create_user_sessions.up.sql`, `models/user_session.go`, `repositories/user_session_repository.go` |

---

## Follow-up: Areas 2, 3, 4.3, 7

| Item | Location |
|------|----------|
| Discord token columns + migration `000012` | `models/user_session.go`, `db/changelog/000012_add_user_sessions_discord_tokens.up.sql` |
| `FindByDiscordUserID` | `repositories/user_session_repository.go` |
| AES-GCM session encryption | `auth/encryption.go`, `auth/encryption_test.go` |
| PKCE (default on; disable with `OAUTH2_PKCE_ENABLED=false`) | `auth/pkce.go`, `auth/oauth2.go`, `handlers/api/routeauth/auth_handlers.go` |
| Encrypted Discord tokens on OAuth callback | `handlers/api/routeauth/auth_handlers.go` |
| Bearer `updated_by_id` on `POST /groceries` | `handlers/api/routes.go` |

Env for `/auth/*`: `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, `DISCORD_REDIRECT_URI`, **`SESSION_ENCRYPTION_KEY`** (hex 32-byte key; OAuth routes omitted if missing/invalid). Optional: `APP_REDIRECT_URI`, `OAUTH2_PKCE_ENABLED`. `JWT_SIGNING_KEY` required for JWT issue/verify.

---

## Status Overview

### Done

| # | Item | Location | AC |
|---|------|----------|----|
| 1 | JWT issuing + verification (HS256, 15-min TTL) | `auth/user_access_jwt.go` | AC-2 |
| 2 | Bearer auth in middleware (verifies JWT, extracts `sub`) | `handlers/api/middleware/auth_middleware.go:115-160` | AC-2 |
| 3 | `X-Guild-ID` header handling + guild membership check (bot session) | `handlers/api/middleware/auth_middleware.go:136-154` | AC-5 |
| 4 | CORS allows `X-Guild-ID` header | `handlers/api/routes.go:50-51` | AC-7 |
| 5 | Rate limiter keys Bearer requests by Discord user ID | `handlers/api/middleware/auth_middleware.go:52-57, 124-125` | AC-8 |
| 6 | Basic auth backward compatibility preserved | `handlers/api/middleware/auth_middleware.go:76-114` | AC-7 |
| 7 | `/metrics` remains unauthenticated | `handlers/api/middleware/auth_middleware.go:63-66` | AC-7 |
| 8 | Middleware skip for `/guilds` path (no `X-Guild-ID` required) | `handlers/api/middleware/auth_middleware.go:41-43` | AC-4 |
| 9 | Local-only test JWT endpoint (`POST /.test/issue-jwt`) | `handlers/api/routes.go` | — |
| 10 | `golang.org/x/oauth2` direct dependency | `go.mod` | AC-10 |
| 11 | Discord OAuth2 env → `oauth2.Config` + `APP_REDIRECT_URI` + PKCE flag + `TokenEncryptor` | `auth/oauth2.go` | AC-10 |
| 12 | Skip `/auth/*` when OAuth/session encryption env incomplete; no panic | `handlers/api/routes.go`, `auth/oauth2.go` | AC-10 |
| 13 | `GET /auth/discord` (state + redirect; PKCE when enabled) | `handlers/api/routeauth/auth_handlers.go` | AC-1 |
| 14 | `GET /auth/discord/callback` (exchange, `@me`, JWT + refresh, encrypt/store Discord tokens, redirect) | `handlers/api/routeauth/auth_handlers.go` | AC-1 |
| 15 | Refresh token issue (random), SHA-256 storage, 7-day TTL | `auth/refresh_token.go`, `user_sessions` | AC-3 |
| 16 | `POST /auth/refresh` (hash lookup, rotate refresh, new JWT) | `handlers/api/routeauth/auth_handlers.go` | AC-3 |
| 17 | `user_sessions` migrations `000011` + `000012`, model, repository | `db/changelog/`, `models/user_session.go`, `repositories/user_session_repository.go` | AC-9 (Discord token columns done) |
| 18 | Unauthenticated `/auth/*` (middleware prefix skip) | `handlers/api/middleware/auth_middleware.go` | — |
| 19 | AES-GCM `SESSION_ENCRYPTION_KEY`; Discord tokens encrypted at rest | `auth/encryption.go` | AC-1 |
| 20 | PKCE (`code_verifier` / S256 challenge) | `auth/pkce.go`, `handlers/api/routeauth/auth_handlers.go` | — |
| 21 | `POST /groceries` sets `UpdatedByID` from JWT when Bearer | `handlers/api/routes.go` | AC-6 |

### Missing

Items are grouped by area and listed in suggested implementation order (dependencies flow top-down).

#### Area 1: Configuration & Dependencies (AC-10)

- [x] **1.1** Add `golang.org/x/oauth2` as a direct dependency in `go.mod`.
- [x] **1.2** Read `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, `DISCORD_REDIRECT_URI` from env and build an `oauth2.Config` with Discord endpoints.
- [x] **1.3** Add a configurable app redirect URI (e.g. env var `APP_REDIRECT_URI`, default `grocerybot://auth/callback`) for the final redirect back to the mobile app after callback.
- [x] **1.4** Graceful degradation: if any Discord OAuth2 env vars are missing, skip registering `/auth/*` routes. The bot must not panic; Basic auth continues to work. (`/guilds` not registered yet — still missing Area 6.) **`SESSION_ENCRYPTION_KEY` is also required for `/auth/*`.**

#### Area 2: Database — `user_sessions` Table (AC-9)

- [x] **2.1** Extend model with fields: `EncryptedDiscordAccessToken`, `EncryptedDiscordRefreshToken`, `DiscordTokenExpiry` (and keep existing refresh columns).
- [x] **2.2** Follow-up migration altering `user_sessions` for Discord token columns (initial `000011` creates minimal table only).
- [x] **2.3** Repository `repositories/user_session_repository.go` with context-aware methods (minimal schema + `FindByDiscordUserID`).

#### Area 3: Encryption for Discord Tokens (AC-1)

- [x] **3.1** Implement AES-GCM encrypt/decrypt utility (`auth/encryption.go`) keyed by env var `SESSION_ENCRYPTION_KEY`.
- [x] **3.2** Use the utility when storing Discord access and refresh tokens in `user_sessions` (OAuth callback).

#### Area 4: OAuth2 Routes (AC-1)

- [x] **4.1** `GET /auth/discord` — generate random `state`, store in short-lived cache, redirect to Discord authorization URL.
- [x] **4.2** `GET /auth/discord/callback` — validate `state`, exchange `code`, `GET /users/@me`, issue GroceryBot JWT + refresh token, redirect to `APP_REDIRECT_URI`. Stores encrypted Discord tokens.
- [x] **4.3** PKCE support (generate `code_verifier`/`code_challenge`, pass to Discord, use in exchange). Default on; `OAUTH2_PKCE_ENABLED=false` to disable.

#### Area 5: Refresh Tokens (AC-3)

- [x] **5.1** Generate cryptographically random refresh token on login.
- [x] **5.2** Store SHA-256 hash of refresh token in `user_sessions` with 7-day expiry.
- [x] **5.3** `POST /auth/refresh` — accept refresh token in body, look up by hash, check expiry, issue new JWT; **rotates** refresh token.
- [x] **5.4** Return `401` if refresh token is expired/missing, forcing re-auth with Discord.

#### Area 6: Guild Listing (AC-4)

- [ ] **6.1** `GET /guilds` handler — load stored Discord access token for the authenticated user, call Discord `GET /users/@me/guilds`, intersect with `discordSess.State.Guilds`, return `[{id, name, icon}]`.
- [ ] **6.2** If stored Discord token is expired, refresh it using the stored Discord refresh token before retrying.
- [ ] **6.3** If Discord token cannot be refreshed (revoked), return `401` indicating re-auth is required.

#### Area 7: User Attribution (AC-6)

- [x] **7.1** In `POST /groceries` handler (`handlers/api/routes.go`), when `authContext.UserID != ""`, set `groceryEntry.UpdatedByID = &authContext.UserID` server-side, overriding any client-supplied value.
- [ ] **7.2** Apply the same pattern to future edit/delete endpoints.

---

## Notes

- The `auth_middleware.go` Bearer path already uses `discordSess.Guild()` and `discordSess.GuildMember()` for per-request guild membership checks with a 60-second cache. This satisfies AC-5 without needing the stored Discord token for CRUD requests.
- The stored Discord tokens (Area 3) are required for `GET /guilds` (Area 6), which calls Discord's REST API on behalf of the user.
- PKCE is on by default; set `OAUTH2_PKCE_ENABLED=false` only if needed for debugging or a legacy client.
