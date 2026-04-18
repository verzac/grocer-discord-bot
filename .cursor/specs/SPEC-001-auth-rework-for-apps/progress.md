# SPEC-001 Progress Tracker

> Last updated: 2026-04-18

Reference spec: `.cursor/specs/SPEC-001-auth-rework-for-apps/spec.md`

---

## Handover Summary (read this first)

**Current state of the codebase:** **Option B (Client-Driven Flow)** is implemented for the OAuth login flow. The Expo app drives OAuth with Discord (`expo-auth-session`); the backend exposes `POST /auth/token` (JSON tokens), `POST /auth/refresh` (includes `expires_in`), and `POST /auth/logout` (Bearer JWT — middleware carve-out so only this `/auth/*` route is authenticated).

**Decision made:** Option B (chosen). Option A routes and env vars have been removed.

**Why:** Option A stored state/PKCE in an in-memory cache (lost on restart), put tokens in URL query params, and bypassed Expo's auth primitives. Option B is stateless on the server during login, returns tokens via JSON, and is more idiomatic for Expo.

**Area 4 (tasks 4.4–4.8):** **Done.**

**Logout (4.5):** Uses **Bearer JWT** only (Approach B). `POST /auth/logout` is excluded from the `/auth/*` auth skip in middleware so the JWT is verified; `X-Guild-ID` is not required (same pattern as `/guilds`). Handler deletes `user_sessions` rows for `sub`.

Also still outstanding: **Area 6** (`GET /guilds`) and **Area 7.2** (user attribution on future edit/delete endpoints).

**What to keep unchanged:** JWT issuing/verification (`auth/user_access_jwt.go`), Bearer middleware, `POST /auth/refresh`, encryption (`auth/encryption.go`), refresh token logic, `user_sessions` table + model + repo, `golang.org/x/oauth2` (still used for code exchange in the new endpoint).

**Key implementation detail:** In Option A, `oauth2.Config.RedirectURL` was set once from `DISCORD_REDIRECT_URI`. In Option B, the redirect URI comes from the client per request. Set `oauth2.Config.RedirectURL` to empty and pass the client-supplied (validated) redirect URI via `oauth2.SetAuthURLParam("redirect_uri", clientRedirectURI)` in the `Exchange()` call.

---

## Decision: OAuth2 Flow Architecture

### ~~Option A — Backend Callback~~ (superseded)

The Discord OAuth2 redirect URI points to the **GroceryBot backend** (`api.grocerybot.net/auth/discord/callback`), not the app directly. After exchanging the code for tokens and issuing a GroceryBot JWT, the backend redirects the browser to the app via a custom scheme (`grocerybot://auth/callback?access_token=...&refresh_token=...`). This keeps the `DISCORD_CLIENT_SECRET` server-side and simplifies redirect URI registration with Discord.

### Option B — Client-Driven Flow (chosen)

The Expo app drives the OAuth2 Authorization Code + PKCE flow directly with Discord. The backend receives the authorization code (and PKCE verifier) from the client and performs the secure exchange.

**Flow:**

1. Expo app uses `expo-auth-session` (`useAuthRequest`) to generate `state` + PKCE (`code_verifier`/`code_challenge`) client-side.
2. Opens system browser to Discord's authorization URL directly (not via the backend).
3. Discord redirects back to the app via custom scheme (`grocerybot://auth/callback?code=...`).
4. `expo-auth-session` parses the authorization `code` from the redirect.
5. App POSTs `{ code, code_verifier, redirect_uri }` to `POST /auth/token` on the backend.
6. Backend exchanges the code + verifier with Discord (using `DISCORD_CLIENT_SECRET`), calls `GET /users/@me`, stores encrypted Discord tokens, issues GroceryBot JWT + refresh token.
7. Backend returns `{ access_token, refresh_token, expires_in }` as JSON.

**Why Option B over Option A:**

- **Eliminates in-memory state cache on the server.** Option A stores `state` + PKCE verifier in `go-cache` (in-memory, lost on restart). Option B moves state/PKCE management to the client via `expo-auth-session` — the server is stateless during the login flow.
- **Tokens returned via JSON, not URL query params.** Cleaner, more standard, and avoids tokens in redirect URLs.
- **Richer client-side error handling.** `expo-auth-session` provides typed response states (`success`, `error`, `dismiss`). The app can distinguish user cancellation from provider errors before making any backend call.
- **Returns `expires_in` naturally.** The JSON response includes token expiry, so the Expo app doesn't need to decode the JWT or hardcode the TTL.
- **More Expo-idiomatic.** Follows the patterns in Expo's official docs and community examples. Easier to find help and benefit from future library improvements.
- **No production users on Option A.** No migration concern — we can refactor freely.

**What changes from Option A:**

| Aspect | Option A (old) | Option B (new) |
|--------|---------------|----------------|
| Who generates `state` + PKCE | Backend (`go-cache`) | Expo client (`expo-auth-session`) |
| Discord redirect URI | Backend: `api.grocerybot.net/auth/discord/callback` | App custom scheme: `grocerybot://auth/callback` |
| `GET /auth/discord` | Generates state, redirects to Discord | **Removed** |
| `GET /auth/discord/callback` | Exchanges code, redirects to app with tokens | **Removed** |
| New: `POST /auth/token` | N/A | Accepts `{ code, code_verifier, redirect_uri }`, returns JSON `{ access_token, refresh_token, expires_in }` |
| `POST /auth/refresh` | Returns access + refresh only | Returns `expires_in` (seconds) for the access JWT |
| `DISCORD_REDIRECT_URI` env var | Points to backend callback URL | **Removed** — redirect URI comes from the client in `POST /auth/token` (validated against an allowlist) |
| `APP_REDIRECT_URI` env var | Where backend redirects after callback | **Removed** — the app handles the redirect itself |
| `OAUTH2_PKCE_ENABLED` env var | Toggles server-side PKCE | **Removed** — PKCE is always on, managed by the client |
| Discord Developer Portal redirect URIs | `https://api.grocerybot.net/auth/discord/callback` | `grocerybot://auth/callback` (custom scheme) |

**Security considerations:**

- `DISCORD_CLIENT_SECRET` remains server-side — the client only sends the authorization code and PKCE verifier, never the secret.
- The backend must validate `redirect_uri` in `POST /auth/token` against an allowlist to prevent code injection with an attacker-controlled redirect. Use env var `ALLOWED_REDIRECT_URIS` (comma-separated) or hardcode the expected scheme.
- PKCE (S256) is mandatory for public clients per RFC 7636. `expo-auth-session` enables it by default.

**Expo libraries needed:**

- `expo-auth-session` — manages PKCE, state, authorization URL, parses code from redirect
- `expo-web-browser` — peer dependency, opens system browser for auth
- `expo-secure-store` — persist JWT and refresh token securely on-device
- `expo-linking` — handles deep link / custom scheme registration

**New endpoints:**

- `POST /auth/token` — exchange authorization code for GroceryBot tokens (replaces `GET /auth/discord` + `GET /auth/discord/callback`)
- `POST /auth/logout` — revoke session (new; not present in Option A either)

**Env var changes:**

- **Keep:** `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, `SESSION_ENCRYPTION_KEY`, `JWT_SIGNING_KEY`
- **Add:** `ALLOWED_REDIRECT_URIS` (comma-separated allowlist for `redirect_uri` validation in `POST /auth/token`)
- **Remove:** `DISCORD_REDIRECT_URI`, `APP_REDIRECT_URI`, `OAUTH2_PKCE_ENABLED`

---

## Thin slice: Areas 1, 4, 5 (historical)

Implemented first: Discord OAuth2 login → GroceryBot JWT + refresh token; refresh endpoint. Discord tokens and PKCE were added in a follow-up (see below).

| Item | Location |
|------|----------|
| OAuth2 config (historical; Option B uses `ALLOWED_REDIRECT_URIS`) | `auth/oauth2.go` |
| Refresh token generate + SHA-256 hash | `auth/refresh_token.go` |
| `POST /auth/token`, `POST /auth/refresh`, `POST /auth/logout` (current) | `handlers/api/routeauth/auth_handlers.go` |
| Register auth routes when Discord env complete | `handlers/api/routes.go` |
| Skip auth for `/auth/*` except `/auth/logout` (Bearer) | `handlers/api/middleware/auth_middleware.go` |
| Minimal `user_sessions` migration + model + repo | `db/changelog/000011_create_user_sessions.up.sql`, `models/user_session.go`, `repositories/user_session_repository.go` |

---

## Follow-up: Areas 2, 3, 4.3, 7

| Item | Location |
|------|----------|
| Discord token columns + migration `000012` | `models/user_session.go`, `db/changelog/000012_add_user_sessions_discord_tokens.up.sql` |
| `FindByDiscordUserID` | `repositories/user_session_repository.go` |
| AES-GCM session encryption | `auth/encryption.go`, `auth/encryption_test.go` |
| PKCE (client-side for Option B; `auth/pkce.go` retained as utility) | `auth/pkce.go` |
| Encrypted Discord tokens on token exchange (`POST /auth/token`) | `handlers/api/routeauth/auth_handlers.go` |
| Bearer `updated_by_id` on `POST /groceries` | `handlers/api/routes.go` |

Env for `/auth/*`: `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, **`SESSION_ENCRYPTION_KEY`** (hex 32-byte key; OAuth routes omitted if missing/invalid). Optional: **`ALLOWED_REDIRECT_URIS`** (comma-separated; defaults to `grocerybot://auth/callback`). `JWT_SIGNING_KEY` required for JWT issue/verify. **`DISCORD_REDIRECT_URI` / `APP_REDIRECT_URI` / `OAUTH2_PKCE_ENABLED` removed** (Option B).

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
| 7 | `/metrics` remains unauthenticated | `handlers/api/middleware/auth_middleware.go` | AC-7 |
| 8 | Middleware skip for `/guilds` path (no `X-Guild-ID` required) | `handlers/api/middleware/auth_middleware.go` | AC-4 |
| 9 | Local-only test JWT endpoint (`POST /.test/issue-jwt`) | `handlers/api/routes.go` | — |
| 10 | `golang.org/x/oauth2` direct dependency | `go.mod` | AC-10 |
| 11 | Discord OAuth2 env → `oauth2.Config` + `ALLOWED_REDIRECT_URIS` + `TokenEncryptor` (empty `RedirectURL`; per-request in exchange) | `auth/oauth2.go` | AC-10 |
| 12 | Skip `/auth/*` when OAuth/session encryption env incomplete; no panic | `handlers/api/routes.go`, `auth/oauth2.go` | AC-10 |
| 13 | `POST /auth/token` (allowlist `redirect_uri`, exchange, `@me`, JWT + refresh, encrypt/store Discord tokens, JSON incl. `expires_in`) | `handlers/api/routeauth/auth_handlers.go` | AC-1 |
| 14 | `POST /auth/logout` (Bearer JWT; delete sessions by `sub`; middleware carve-out) | `handlers/api/routeauth/auth_handlers.go` | — |
| 15 | Refresh token issue (random), SHA-256 storage, 7-day TTL | `auth/refresh_token.go`, `user_sessions` | AC-3 |
| 16 | `POST /auth/refresh` (hash lookup, rotate refresh, new JWT, **`expires_in`**) | `handlers/api/routeauth/auth_handlers.go` | AC-3 |
| 17 | `user_sessions` migrations `000011` + `000012`, model, repository | `db/changelog/`, `models/user_session.go`, `repositories/user_session_repository.go` | AC-9 (Discord token columns done) |
| 18 | Unauthenticated `/auth/*` except `/auth/logout` (Bearer; no `X-Guild-ID`) | `handlers/api/middleware/auth_middleware.go` | — |
| 19 | AES-GCM `SESSION_ENCRYPTION_KEY`; Discord tokens encrypted at rest | `auth/encryption.go` | AC-1 |
| 20 | PKCE client-side (Option B); `auth/pkce.go` utility only | `auth/pkce.go` | — |
| 21 | `POST /groceries` sets `UpdatedByID` from JWT when Bearer | `handlers/api/routes.go` | AC-6 |

**Option B refactor (tasks 4.4–4.8):** **Completed.** Old items #13/#14 (`GET /auth/discord`, callback) removed; #11 refactored; server-side PKCE initiation removed from `auth_handlers.go`.

### Missing

Items are grouped by area and listed in suggested implementation order (dependencies flow top-down).

#### Area 1: Configuration & Dependencies (AC-10)

- [x] **1.1** Add `golang.org/x/oauth2` as a direct dependency in `go.mod`.
- [x] **1.2** Read `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET` from env and build an `oauth2.Config` with Discord endpoints. (Option B: `RedirectURL` empty; `DISCORD_REDIRECT_URI` removed.)
- [x] **1.3** Redirect URI allowlist for token exchange: `ALLOWED_REDIRECT_URIS` (comma-separated; default `grocerybot://auth/callback`). Replaces `APP_REDIRECT_URI` post-callback redirect.
- [x] **1.4** Graceful degradation: if any Discord OAuth2 env vars are missing, skip registering `/auth/*` routes. The bot must not panic; Basic auth continues to work. (`/guilds` not registered yet — still missing Area 6.) **`SESSION_ENCRYPTION_KEY` is also required for `/auth/*`.**

#### Area 2: Database — `user_sessions` Table (AC-9)

- [x] **2.1** Extend model with fields: `EncryptedDiscordAccessToken`, `EncryptedDiscordRefreshToken`, `DiscordTokenExpiry` (and keep existing refresh columns).
- [x] **2.2** Follow-up migration altering `user_sessions` for Discord token columns (initial `000011` creates minimal table only).
- [x] **2.3** Repository `repositories/user_session_repository.go` with context-aware methods (minimal schema + `FindByDiscordUserID`).

#### Area 3: Encryption for Discord Tokens (AC-1)

- [x] **3.1** Implement AES-GCM encrypt/decrypt utility (`auth/encryption.go`) keyed by env var `SESSION_ENCRYPTION_KEY`.
- [x] **3.2** Use the utility when storing Discord access and refresh tokens in `user_sessions` (OAuth callback).

#### Area 4: OAuth2 Routes (AC-1) — refactored for Option B

- [x] ~~**4.1** `GET /auth/discord` — generate random `state`, store in short-lived cache, redirect to Discord authorization URL.~~ **Removed** (client drives the flow now).
- [x] ~~**4.2** `GET /auth/discord/callback` — validate `state`, exchange `code`, `GET /users/@me`, issue GroceryBot JWT + refresh token, redirect to `APP_REDIRECT_URI`. Stores encrypted Discord tokens.~~ **Removed** (replaced by `POST /auth/token`).
- [x] ~~**4.3** PKCE support (generate `code_verifier`/`code_challenge`, pass to Discord, use in exchange). Default on; `OAUTH2_PKCE_ENABLED=false` to disable.~~ **Removed** (PKCE is now client-side via `expo-auth-session`; server receives verifier in `POST /auth/token`).
- [x] **4.4** `POST /auth/token` — accept `{ code, code_verifier, redirect_uri }`, validate `redirect_uri` against `ALLOWED_REDIRECT_URIS`, exchange code + verifier with Discord using `DISCORD_CLIENT_SECRET`, call `GET /users/@me`, encrypt/store Discord tokens, issue GroceryBot JWT + refresh token, return `{ access_token, refresh_token, expires_in }` as JSON.
- [x] **4.5** `POST /auth/logout` — **Bearer JWT required** (middleware carve-out); delete the user's session(s) from `user_sessions` by Discord user ID (`sub`); return 204.
- [x] **4.6** Remove `GET /auth/discord`, `GET /auth/discord/callback`, in-memory `stateCache`, server-side PKCE generation from `auth_handlers.go`.
- [x] **4.7** Remove `APP_REDIRECT_URI` and `OAUTH2_PKCE_ENABLED` from `auth/oauth2.go`. Replace `DISCORD_REDIRECT_URI` with `ALLOWED_REDIRECT_URIS`.
- [x] **4.8** Add `expires_in` field to `POST /auth/refresh` JSON response.

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
- ~~PKCE is on by default; set `OAUTH2_PKCE_ENABLED=false` only if needed for debugging or a legacy client.~~ PKCE is now client-side (Option B); the server always expects `code_verifier` in `POST /auth/token`.

### Option B refactor notes

- **What to keep:** JWT issuing/verification, Bearer middleware, `POST /auth/refresh`, encryption, refresh token logic, `user_sessions` table + model + repo, `auth/pkce.go` (useful reference but no longer called server-side for initiation). The `oauth2.Config` from `golang.org/x/oauth2` is still used for the code exchange in `POST /auth/token`.
- **What to remove/replace:** `GET /auth/discord`, `GET /auth/discord/callback`, in-memory `stateCache`, `oauthStateEntry` struct, server-side PKCE verifier generation during initiation, `APP_REDIRECT_URI`, `OAUTH2_PKCE_ENABLED` env var, `DISCORD_REDIRECT_URI` env var.
- **What to add:** `POST /auth/token`, `POST /auth/logout` (**Bearer** + middleware carve-out), `ALLOWED_REDIRECT_URIS` env var, `expires_in` in token and refresh responses.
- **`redirect_uri` validation:** The `POST /auth/token` endpoint must validate the client-supplied `redirect_uri` against `ALLOWED_REDIRECT_URIS`. This is critical — without it, an attacker could exchange a code obtained via a malicious redirect URI. The allowlist should contain the app's custom scheme (e.g. `grocerybot://auth/callback`).
- **`oauth2.Config.RedirectURL`:** In Option A this was set once from `DISCORD_REDIRECT_URI`. In Option B, the redirect URL varies per request (the client sends it). The `Exchange()` call must pass it via `oauth2.SetAuthURLParam("redirect_uri", clientRedirectURI)` so Discord can verify it matches the one used in the authorization request.
- **Discord Developer Portal:** Must register `grocerybot://auth/callback` (the custom scheme) as an allowed redirect URI. The old backend callback URL can be removed.
