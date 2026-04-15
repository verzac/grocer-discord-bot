# SPEC-001 Progress Tracker

> Last updated: 2026-04-16

Reference spec: `.cursor/specs/SPEC-001-auth-rework-for-apps/spec.md`

---

## Decision: OAuth2 Flow Architecture (Option A — Backend Callback)

The Discord OAuth2 redirect URI points to the **GroceryBot backend** (`api.grocerybot.net/auth/discord/callback`), not the app directly. After exchanging the code for tokens and issuing a GroceryBot JWT, the backend redirects the browser to the app via a custom scheme (`grocerybot://auth/callback?access_token=...&refresh_token=...`). This keeps the `DISCORD_CLIENT_SECRET` server-side and simplifies redirect URI registration with Discord.

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
| 9 | Local-only test JWT endpoint (`POST /.test/issue-jwt`) | `handlers/api/routes.go:66-85` | — |

### Missing

Items are grouped by area and listed in suggested implementation order (dependencies flow top-down).

#### Area 1: Configuration & Dependencies (AC-10)

- [ ] **1.1** Add `golang.org/x/oauth2` as a direct dependency in `go.mod`.
- [ ] **1.2** Read `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, `DISCORD_REDIRECT_URI` from env and build an `oauth2.Config` with Discord endpoints.
- [ ] **1.3** Add a configurable app redirect URI (e.g. env var `APP_REDIRECT_URI`, default `grocerybot://auth/callback`) for the final redirect back to the mobile app after callback.
- [ ] **1.4** Graceful degradation: if any Discord OAuth2 env vars are missing, skip registering `/auth/*` and `/guilds` routes. The bot must not panic; Basic auth continues to work.

#### Area 2: Database — `user_sessions` Table (AC-9)

- [ ] **2.1** Create model `models/user_session.go` with fields: `ID`, `DiscordUserID`, `EncryptedDiscordAccessToken`, `EncryptedDiscordRefreshToken`, `DiscordTokenExpiry`, `RefreshTokenHash`, `RefreshTokenExpiry`, `CreatedAt`, `UpdatedAt`.
- [ ] **2.2** Create migration `db/changelog/000011_create_user_sessions.up.sql` (and `.down.sql`) to create the `user_sessions` table.
- [ ] **2.3** Create repository `repositories/user_session_repository.go` with CRUD methods (accepting `context.Context`).

#### Area 3: Encryption for Discord Tokens (AC-1)

- [ ] **3.1** Implement AES-GCM encrypt/decrypt utility (e.g. `auth/encryption.go`) keyed by a new env var `SESSION_ENCRYPTION_KEY`.
- [ ] **3.2** Use the utility when storing/retrieving Discord access and refresh tokens in `user_sessions`.

#### Area 4: OAuth2 Routes (AC-1)

- [ ] **4.1** `GET /auth/discord` — generate random `state` (+ optional PKCE `code_verifier`), store in short-lived cache, redirect to Discord authorization URL.
- [ ] **4.2** `GET /auth/discord/callback` — validate `state`, exchange `code` for Discord tokens via `oauth2.Config.Exchange()`, call Discord `GET /users/@me` to get user ID, store encrypted Discord tokens in `user_sessions`, issue GroceryBot JWT + refresh token, redirect browser to `APP_REDIRECT_URI?access_token=...&refresh_token=...`.
- [ ] **4.3** PKCE support for public clients (generate `code_verifier`/`code_challenge`, pass to Discord, use in exchange). Discord supports PKCE.

#### Area 5: Refresh Tokens (AC-3)

- [ ] **5.1** Generate cryptographically random refresh token on login.
- [ ] **5.2** Store SHA-256 hash of refresh token in `user_sessions` with 7-day expiry.
- [ ] **5.3** `POST /auth/refresh` — accept refresh token in body, look up by hash, check expiry, issue new JWT. Optionally rotate refresh token (issue new, invalidate old).
- [ ] **5.4** Return `401` if refresh token is expired/missing, forcing re-auth with Discord.

#### Area 6: Guild Listing (AC-4)

- [ ] **6.1** `GET /guilds` handler — load stored Discord access token for the authenticated user, call Discord `GET /users/@me/guilds`, intersect with `discordSess.State.Guilds`, return `[{id, name, icon}]`.
- [ ] **6.2** If stored Discord token is expired, refresh it using the stored Discord refresh token before retrying.
- [ ] **6.3** If Discord token cannot be refreshed (revoked), return `401` indicating re-auth is required.

#### Area 7: User Attribution (AC-6)

- [ ] **7.1** In `POST /groceries` handler (`handlers/api/routes.go`), when `authContext.UserID != ""`, set `groceryEntry.UpdatedByID = &authContext.UserID` server-side, overriding any client-supplied value.
- [ ] **7.2** Apply the same pattern to future edit/delete endpoints.

---

## Notes

- The `auth_middleware.go` Bearer path already uses `discordSess.Guild()` and `discordSess.GuildMember()` for per-request guild membership checks with a 60-second cache. This satisfies AC-5 without needing the stored Discord token for CRUD requests.
- The stored Discord tokens (Area 3) are only needed for `GET /guilds` (Area 6), which calls Discord's REST API on behalf of the user.
- PKCE (Area 4.3) is recommended but could be deferred to a follow-up if the initial client is a trusted first-party app.
