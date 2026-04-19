# GroceryBot API Externalization

## 1. Background & Motivation

GroceryBot's REST API currently uses **guild-scoped Basic auth**. A Discord server administrator runs `/developer` in Discord to generate a `client_id` (set to the guild ID) and a `client_secret` (random 32-char string, bcrypt-hashed in DB). Every API request carries `Authorization: Basic base64(clientID:clientSecret)`, and the middleware resolves the target guild from the `api_clients.scope` field (`guild:<guildID>`).

This model works for server-admin integrations but **cannot support end-user-facing apps** (mobile or web) because:

- Credentials are per-guild, not per-user — a single leaked key exposes the entire guild's data.
- There is no way for a user to authenticate as *themselves* and browse across multiple guilds.
- Public clients (SPAs, mobile apps) cannot safely store long-lived shared secrets.

The goal is to let an individual Discord user:

1. **Log in via Discord** (from a mobile app or web app).
2. **Select a guild** they belong to (that also has GroceryBot installed).
3. **CRUD groceries** on that guild's lists through the API.

## 2. Current State

### 2.1 API Routes

| Method   | Path               | Purpose                          |
|----------|--------------------|----------------------------------|
| `GET`    | `/grocery-lists`   | List grocery entries & lists     |
| `POST`   | `/groceries`       | Create a grocery entry           |
| `DELETE`  | `/groceries/:id`   | Delete a grocery entry           |
| `GET`    | `/registrations`   | List guild registrations         |
| `GET`    | `/metrics`         | Prometheus metrics (no auth)     |

Some operations are **not yet exposed** via the API (e.g. editing a grocery entry, managing grocery lists, clearing lists). This is acceptable — the spec focuses on the auth and guild-selection layers; CRUD coverage can be expanded later.

### 2.2 Auth Flow (Current)

```
Discord Admin runs /developer
  → bot generates clientID (= guildID), clientSecret (random)
  → stores bcrypt(clientSecret) + scope "guild:<guildID>" in api_clients
  → returns plaintext credentials once (ephemeral)

API consumer sends:  Authorization: Basic base64(guildID:clientSecret)
  → AuthMiddleware decodes, loads api_clients row by client_id
  → bcrypt.CompareHashAndPassword
  → wraps echo.Context in AuthContext{Scope: row.Scope}
  → handler calls auth.GetGuildIDFromScope(scope) to resolve guild
```

### 2.3 Data Model (Relevant)

- **`api_clients`** — `id`, `client_id`, `client_secret` (bcrypt), `scope`, `created_by_id`, `created_at`, `updated_at`, `deleted_at`.
- **`grocery_entries`** — `id`, `item_desc`, `guild_id`, `grocery_list_id` (FK), `updated_by_id`, timestamps.
- **`grocery_lists`** — `id`, `guild_id`, `list_label`, `fancy_name`, timestamps.
- **`guild_configs`** — `guild_id` (PK), feature flags, timestamps.
- **`guild_registrations`** / **`registration_entitlements`** / **`registration_tiers`** — tier/limit system.

### 2.4 Existing Middleware

CORS (configurable origins), 10 s timeout, rate limiter (10 req / 30 s per client ID or IP), panic recovery, custom validator.

### 2.5 Environment Variables for OAuth2 (Option B)

Used by `auth/oauth2.go` and related handlers when registering `/auth/*` routes. (The hardcoded client ID `815120759680532510` in invite URLs is the bot application ID for the **bot-invite** flow, not the user OAuth2 app config.)

| Variable | Purpose |
|----------|---------|
| `DISCORD_CLIENT_ID` | Discord application client ID for the user OAuth2 Authorization Code + PKCE flow. |
| `DISCORD_CLIENT_SECRET` | Exchanges the authorization code for tokens (server-side only; never sent to the client). |
| `ALLOWED_REDIRECT_URIS` | Comma-separated allowlist of `redirect_uri` values the client may send in **`POST /auth/token`**. Must include every redirect URI registered in the Discord Developer Portal that the app uses (e.g. `grocerybot://auth/callback`, or Expo proxy URL). If unset, defaults to `grocerybot://auth/callback`. |
| `SESSION_ENCRYPTION_KEY` | Hex-encoded 32-byte AES key; required for `/auth/*` registration. Discord tokens encrypted at rest. |
| `JWT_SIGNING_KEY` | HMAC key for GroceryBot access JWTs. |

**Redirect URI:** The client chooses `redirect_uri` when opening Discord's authorize URL; Discord redirects there with `?code=`. That same `redirect_uri` must be registered in the Developer Portal and must be listed in **`ALLOWED_REDIRECT_URIS`** so the backend will accept it in **`POST /auth/token`**. Option B does **not** use a backend callback URL or `DISCORD_REDIRECT_URI`.

Additional configuration is described in §4.2.

---

## 3. Desired User Flow

```
┌─────────────┐      ┌──────────────────┐      ┌────────────────┐
│  Mobile /    │      │  GroceryBot API  │      │  Discord API   │
│  Web App     │      │  (Echo :8080)    │      │                │
└──────┬──────┘      └────────┬─────────┘      └───────┬────────┘
       │                      │                        │
       │  1. "Log in with Discord"                     │
       │──────────────────────────────────────────────►│
       │  2. OAuth2 redirect / token                   │
       │◄──────────────────────────────────────────────│
       │                      │                        │
       │  3. Present token    │                        │
       │─────────────────────►│                        │
       │                      │  4. Validate user &    │
       │                      │     fetch guilds       │
       │                      │───────────────────────►│
       │                      │◄───────────────────────│
       │  5. Return guild list│                        │
       │◄─────────────────────│                        │
       │                      │                        │
       │  6. Select guild     │                        │
       │─────────────────────►│                        │
       │                      │                        │
       │  7. CRUD groceries   │                        │
       │─────────────────────►│                        │
       │  8. Response         │                        │
       │◄─────────────────────│                        │
```

**Key requirements:**

- **R1.** User authenticates as a Discord user (not as a guild).
- **R2.** User can enumerate guilds where (a) they are a member AND (b) GroceryBot is installed.
- **R3.** User selects a guild, then CRUDs groceries scoped to that guild.
- **R4.** Existing guild-scoped Basic auth continues to work (backward-compatible).
- **R5.** Public clients (SPA / mobile) must not need to store a shared server secret.

---

## 4. Approach — Discord OAuth2 + GroceryBot-Issued JWT

The **mobile/web client** drives the OAuth2 Authorization Code + PKCE flow with Discord; the **GroceryBot backend** performs the confidential code exchange, stores Discord tokens (encrypted), and issues its own short-lived JWTs (**Option B — client-driven flow**).

### 4.1 Flow

1. **Client** (e.g. Expo with `expo-auth-session`) builds Discord's authorization URL with `response_type=code`, `client_id`, `redirect_uri`, `scope=identify guilds`, `state`, and PKCE (`code_challenge` / `code_challenge_method=S256`), and opens the system browser.
2. **Discord** redirects back to the app's **registered** redirect URI (e.g. `grocerybot://auth/callback?code=...&state=...`).
3. **Client** POSTs `{ code, code_verifier, redirect_uri }` to **`POST /auth/token`** on the GroceryBot API. The backend validates `redirect_uri` against **`ALLOWED_REDIRECT_URIS`**, then exchanges the code + verifier with Discord (using `DISCORD_CLIENT_SECRET`).
4. Backend calls Discord `GET /users/@me` to get the user's identity.
5. Backend stores the user's Discord access token (encrypted) and refresh token server-side, associated with the user ID. This is used later for guild lookups.
6. Backend returns JSON **`{ access_token, refresh_token, expires_in }`** (GroceryBot JWT + opaque refresh token). The JWT identifies the user only — **it does not contain a guild list**.
7. Client stores the JWT and refresh token (e.g. secure storage) and includes `Authorization: Bearer <jwt>` on subsequent requests.
8. When the JWT expires, the client calls **`POST /auth/refresh`** with the refresh token; the response includes a new access token, rotated refresh token, and **`expires_in`**.
9. **`POST /auth/logout`** with `Authorization: Bearer <jwt>` deletes the user's session row(s) server-side (middleware treats this path like other Bearer routes except `X-Guild-ID` is not required).
10. When the user wants to select a guild, **`GET /guilds`** (future) will call Discord `GET /users/@me/guilds` using the **stored Discord access token**, intersect with bot-installed guilds, and return a fresh list.
11. CRUD endpoints require an `X-Guild-ID` header (for Bearer-auth requests), except paths explicitly exempt (e.g. `/guilds`, `/auth/logout`). The middleware verifies guild membership with a short-lived cache.

### 4.2 New/Changed Components

| Component | Change |
|-----------|--------|
| **DB** | New `user_sessions` table to store Discord access/refresh tokens (encrypted) keyed by user ID, plus GroceryBot refresh tokens. |
| **Routes** | `POST /auth/token` — accept `{ code, code_verifier, redirect_uri }`; validate `redirect_uri`; exchange with Discord; issue JWT + refresh; return JSON including `expires_in`. |
|  | `POST /auth/logout` — **Bearer JWT required**; delete `user_sessions` for that user; returns **204**. |
|  | `POST /auth/refresh` — exchange a refresh token for a new JWT access token + rotated refresh token; response includes **`expires_in`**. |
|  | `GET /guilds` — fetch the user's guilds **live from Discord** (using stored Discord token), intersect with bot guilds, return fresh list. |
| **Middleware** | Extend `AuthMiddleware` to accept both `Basic` and `Bearer` schemes. Bearer path verifies JWT signature + expiry, extracts user ID. **`/auth/*` is unauthenticated except `/auth/logout`**, which requires Bearer (no `X-Guild-ID`). |
| **Handlers** | Existing CRUD handlers resolve `guildID` from `X-Guild-ID` header (Bearer) or scope (Basic). Middleware verifies user's guild membership (via cached Discord lookup or short-TTL cache). |
| **Config** | `JWT_SIGNING_KEY`. `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`. **`ALLOWED_REDIRECT_URIS`** (comma-separated allowlist for `POST /auth/token`; default `grocerybot://auth/callback`). **`SESSION_ENCRYPTION_KEY`** for Discord tokens at rest. |
| **Dependencies** | `golang.org/x/oauth2` (OAuth2 flow), `github.com/golang-jwt/jwt/v5` (JWT signing/verification). See §4.5. |

### 4.3 Characteristics

**Strengths:**

- **Full control** over token lifetime, claims, and revocation.
- **No per-request Discord API call for CRUD** — JWT is self-contained and verified locally. Discord is only called when the user fetches `GET /guilds` (guild selection screen) or when guild membership needs re-validation (short-TTL cache).
- **Guild list is always fresh at selection time** — because `GET /guilds` calls Discord live using the stored access token, the user always sees their current guilds when choosing one.
- **Works for both confidential and public clients** — the OAuth2 code exchange happens server-side so the client never sees the Discord client secret.
- Supports **PKCE** extension for mobile (public) clients.
- Can embed extra claims (user roles, tier info) to avoid DB lookups on hot paths.

**Trade-offs:**

- Requires encrypted storage of the user's Discord access/refresh tokens server-side (for guild lookups).
- Requires secure storage of `JWT_SIGNING_KEY` and `DISCORD_CLIENT_SECRET`.
- Guild membership check on CRUD requests relies on a short-TTL cache for performance — a very recently removed guild member could still have access for the cache duration (e.g. 60 s).

### 4.4 Refresh Token Flow

Refresh tokens are a **must** — users should not have to re-authenticate with Discord every session. The flow:

1. On login, the backend issues both a **short-lived JWT** (15 min TTL) and a **long-lived refresh token** (7 day TTL) stored in the `user_sessions` table.
2. The client stores both. When the JWT expires, the client calls `POST /auth/refresh` with the refresh token.
3. The backend validates the refresh token (lookup by hash in `user_sessions`, check expiry), then issues a new JWT. Optionally rotates the refresh token (issue a new one, invalidate the old).
4. If the refresh token itself is expired or revoked, the user must re-authenticate with Discord.

This means a user stays logged in for the refresh token's lifetime without ever seeing Discord's OAuth2 page again.

### 4.5 Dependencies

Two standard Go packages handle the OAuth2 + JWT layer:

| Package | Version | Purpose |
|---------|---------|---------|
| `golang.org/x/oauth2` | latest | OAuth2 Authorization Code flow with Discord: code exchange, token refresh, HTTP client injection. |
| `github.com/golang-jwt/jwt/v5` | v5.3.1+ | Sign and verify GroceryBot-issued JWTs (HMAC-SHA256 or similar). |

We self-roll the Discord OAuth2 integration rather than using a managed provider (Auth0) or a multi-provider framework (`goth`, `disgo`). Rationale:

- **Auth0 was evaluated and ruled out.** It supports Discord as a social connection, but retrieving the user's Discord access token (needed for `GET /users/@me/guilds`) requires Auth0's Token Vault + Management API — adding indirection and a runtime dependency for a single-provider use case. It makes sense for multi-provider auth; not here.
- **Higher-level Go libraries** (`markbates/goth`, `disgoorg/disgo/oauth2`) bring session/state abstractions that may conflict with Echo's middleware. None handle the guild-intersection logic we need regardless.
- **Discord is not an OIDC provider** (no `id_token`, no `.well-known/openid-configuration`), so OIDC-focused libraries add no value.
- The total implementation surface is small: state parameter generation, PKCE, callback handler, Discord token storage, refresh logic, and JWT signing middleware — a few hundred lines.

`golang.org/x/oauth2` requires a Discord endpoint config (auth URL + token URL). This is ~5 lines of setup:

```go
// Option B: RedirectURL is empty; pass redirect_uri per Exchange via oauth2.SetAuthURLParam.
var discordOAuthConfig = &oauth2.Config{
    ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
    ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
    RedirectURL:  "",
    Scopes:       []string{"identify", "guilds"},
    Endpoint: oauth2.Endpoint{
        AuthURL:  "https://discord.com/api/oauth2/authorize",
        TokenURL: "https://discord.com/api/oauth2/token",
    },
}
```

---

## 5. Shared Concerns

### 5.1 Guild Resolution & Deployment

**Decision:** The API remains an in-process goroutine within the bot process. However, the bot may be **periodically restarted** for updates, so all session/token state must be **DB-persisted** (SQLite) — nothing stored only in memory survives a restart.

To determine which guilds the bot is installed in (needed by `GET /guilds` and CRUD membership checks):

- **Primary: Bot session state** — `discordgo.Session.State.Guilds` holds the bot's current guild list in memory. Fast and accurate while the process is running.
- **Fallback: DB-derived** — query distinct `guild_id` values from `guild_configs`. Useful if guild resolution ever needs to work outside the bot process, but not required for now.

Since the API is in-process, the bot session state approach is used. The guild list is rebuilt from Discord on every bot startup (discordgo does this automatically), so restarts do not cause stale guild data.

### 5.2 Authorization Within a Guild

**Decision:** Permissive — any guild member can CRUD groceries on that guild. This mirrors the bot's Discord command behavior (no role restrictions on grocery operations; only `/developer` requires Administrator). No Discord role/permission checks are needed for MVP.

Role-based restrictions (e.g. only certain roles can delete) may be added in the future but are out of scope for this work.

### 5.3 User Attribution (`updated_by_id`)

`GroceryEntry.UpdatedByID` stores the Discord user ID of whoever last created or edited the entry. It is used in Discord embeds (grohere) to show "Last updated by @user".

**Current behavior:**

- **Discord commands** set this field server-side from `commandContext.AuthorID` — always correct.
- **Basic auth API** has no concept of a user identity (the `AuthContext` only carries a guild `Scope`). The field is technically bindable from the JSON request body, but there is no server-side enforcement. In practice it is either client-supplied (untrusted) or null.

**With JWT auth:**

The middleware now knows the Discord user ID (from the JWT `sub` claim). The `AuthContext` should be extended to carry this:

```go
type AuthContext struct {
    echo.Context
    Scope   string // Basic auth: guild-scoped (existing)
    UserID  string // Bearer auth: Discord user ID from JWT (new)
    GuildID string // resolved from scope (Basic) or X-Guild-ID header (Bearer)
}
```

CRUD handlers should **override** `UpdatedByID` server-side when a user ID is available, matching what the Discord command handlers already do. This means:

- `POST /groceries` → set `groceryEntry.UpdatedByID = &authContext.UserID`
- Future `PUT /groceries/:id` (edit) → set `entry.UpdatedByID = &authContext.UserID`

For **Basic auth** requests (no user identity), existing behavior is preserved — `UpdatedByID` remains whatever the client sent or null.

This is a net improvement: the field was already there but unused by the API. User-scoped auth fills it in properly for free.

### 5.4 Backward Compatibility

The existing Basic auth flow **must continue to work**. The middleware should support dual schemes:

```
Authorization: Basic ...   → existing api_clients lookup (guild-scoped)
Authorization: Bearer ...  → new user-scoped auth (JWT)
```

No changes to the existing `api_clients` table or `/developer` slash command.

### 5.5 New Endpoint: `GET /guilds`

A new endpoint returns the guilds the authenticated user can interact with:

```json
GET /guilds
Authorization: Bearer <token>

Response 200:
{
  "guilds": [
    { "id": "123456789", "name": "Flatmate's Server", "icon": "abc123" },
    { "id": "987654321", "name": "My Server", "icon": "def456" }
  ]
}
```

This list is the intersection of the user's Discord guilds and the guilds where GroceryBot is installed.

### 5.6 Guild Selection on CRUD Endpoints

Currently, the guild is implicitly derived from the API key's scope. With user-scoped auth (Bearer JWT), the guild must be **explicitly specified** per request via a custom header:

```
X-Guild-ID: 123456789
```

**Decision:** Use the `X-Guild-ID` request header for Bearer-auth requests. This keeps the existing route paths unchanged and separates the guild-selection concern from the URL structure. The middleware reads this header, verifies the user is a member of the specified guild, and makes it available to handlers via `AuthContext.GuildID`.

**Basic-auth requests ignore `X-Guild-ID`** — the guild continues to be derived from `api_clients.scope` as before.

Example request:

```
GET /grocery-lists
Authorization: Bearer <jwt>
X-Guild-ID: 123456789
```

### 5.7 CORS

The allowed origins list (`GROCER_BOT_API_ALLOW_ORIGINS`) will need to include the Expo dev server origin during development (e.g. `http://localhost:19006`) and the production web app domain if applicable. Native React Native requests (not from a web view) do not send an `Origin` header, so CORS does not apply to them — but the `X-Guild-ID` custom header must be added to `AllowHeaders` in the CORS config.

### 5.8 Rate Limiting

The current rate limiter keys on `clientID` (Basic auth) or IP. For Bearer auth, it should key on the Discord user ID to prevent per-user abuse while allowing shared IPs (e.g. NAT).

### 5.9 Discord OAuth2 Scopes Needed

The Discord OAuth2 request needs these scopes:

- `identify` — access to `GET /users/@me` (user ID, username, avatar).
- `guilds` — access to `GET /users/@me/guilds` (list of guilds the user is in).

No bot-level or message-level scopes are required.

### 5.10 PKCE

For public clients (SPA, mobile app), **PKCE (Proof Key for Code Exchange)** should be used with the Authorization Code flow. Discord supports PKCE. This removes the need for the client to hold a client secret.

### 5.11 Mobile Platform

**Decision:** The mobile app will be built with **React Native + Expo**.

This affects the OAuth2 redirect URI. Expo apps use either:

- **Expo AuthSession proxy:** `https://auth.expo.io/@<owner>/<slug>` — Expo's hosted redirect proxy that forwards the authorization code back to the app. Simplest to set up but adds a dependency on Expo's servers.
- **Custom scheme:** e.g. `grocerybot://auth/callback` — a deep link registered in the app. More reliable for production; no Expo proxy dependency.
- **Universal link:** e.g. `https://app.grocerybot.net/auth/callback` — an HTTPS URL that the OS routes to the app. Most robust but requires domain verification.

The chosen redirect URI(s) must be registered in the Discord Developer Portal and listed in **`ALLOWED_REDIRECT_URIS`** on the backend.

The Expo app opens Discord's authorization URL directly, receives the `code` on the deep link, then **`POST /auth/token`** with `{ code, code_verifier, redirect_uri }`. The backend returns **`{ access_token, refresh_token, expires_in }`** as JSON (no tokens in redirect URLs).

### 5.12 Missing CRUD Endpoints

The current API is incomplete. The following operations exist in Discord commands but not in the API:

| Operation | Discord Command | API Endpoint | Status |
|-----------|----------------|--------------|--------|
| Add grocery | `/gro` | `POST /groceries` | Exists |
| Remove grocery | `/groremove` | `DELETE /groceries/:id` | Exists |
| Edit grocery | `/groedit` | — | **Missing** |
| List groceries | `/grolist` | `GET /grocery-lists` | Exists |
| Clear groceries | `/groclear` | — | **Missing** |
| Create grocery list | `/grolist-new` | — | **Missing** |
| Delete grocery list | `/grolist-delete` | — | **Missing** |
| Bulk add | `/grobulk` | — | **Missing** |

These are **not blockers** for the auth externalization work and can be added incrementally. The auth layer should be designed to support them when they arrive.

---

## 6. Acceptance Criteria

### AC-1: Discord OAuth2 Login (Option B)

- [ ] The client drives the authorize URL (PKCE + `state`); **`POST /auth/token`** accepts `{ code, code_verifier, redirect_uri }`, validates `redirect_uri` against **`ALLOWED_REDIRECT_URIS`**, exchanges the code with Discord (including verifier), and calls `GET /users/@me` to resolve the Discord user ID.
- [ ] On success, **`POST /auth/token`** returns JSON: GroceryBot JWT **`access_token`**, **`refresh_token`**, and **`expires_in`** (seconds).
- [ ] The Discord access token and refresh token are stored server-side (encrypted) in the `user_sessions` table, associated with the Discord user ID.
- [ ] If `redirect_uri` is missing or not allowlisted, the API returns **`400`**.
- [ ] If required body fields are missing, the API returns **`400`**.
- [ ] If the Discord code exchange fails, the API returns an appropriate error (e.g. **`502`**).

### AC-1b: Logout

- [ ] **`POST /auth/logout`** requires **`Authorization: Bearer <jwt>`** (middleware carve-out for this path under `/auth/`). **`X-Guild-ID` is not required.**
- [ ] On success, the user's row(s) in **`user_sessions`** are deleted and the API returns **`204`**.

### AC-2: JWT Access Token

- [ ] The JWT contains at minimum: `sub` (Discord user ID), `iat` (issued at), `exp` (expiration).
- [ ] The JWT is signed using `JWT_SIGNING_KEY` (HMAC-SHA256 or equivalent).
- [ ] The JWT has a 15-minute TTL (configurable via env var or constant).
- [ ] API endpoints that require authentication accept `Authorization: Bearer <jwt>` and verify the signature + expiry.
- [ ] An expired or malformed JWT returns `401`.
- [ ] An invalid signature returns `403`.

### AC-3: Refresh Token

- [ ] On login, a long-lived refresh token is issued alongside the JWT and stored (hashed) in the `user_sessions` table with a 7-day expiry.
- [ ] `POST /auth/refresh` accepts a refresh token in the request body, validates it (hash lookup + expiry check), and returns a new JWT access token, a rotated refresh token, and **`expires_in`** (seconds) for the access token.
- [ ] If the refresh token is expired or not found, `POST /auth/refresh` returns `401`, forcing the user to re-authenticate with Discord.
- [ ] A user can stay logged in for the refresh token's lifetime without re-authenticating with Discord.

### AC-4: Guild Listing

- [ ] `GET /guilds` is accessible only with a valid Bearer JWT.
- [ ] The endpoint calls Discord `GET /users/@me/guilds` using the stored Discord access token for the authenticated user.
- [ ] The returned list is the intersection of the user's Discord guilds and the guilds where GroceryBot is installed.
- [ ] Each guild object in the response includes at minimum: `id`, `name`, `icon`.
- [ ] If the stored Discord token is expired, the backend refreshes it using the stored Discord refresh token before retrying the guild fetch.
- [ ] If the Discord token cannot be refreshed (e.g. user revoked access), the endpoint returns `401` with a message indicating re-authentication is required.

### AC-5: Guild-Scoped CRUD with Bearer Auth

- [ ] Existing CRUD endpoints (`GET /grocery-lists`, `POST /groceries`, `DELETE /groceries/:id`) require the `X-Guild-ID` request header when authenticated via Bearer JWT.
- [ ] The middleware reads `X-Guild-ID`, verifies that the authenticated user is a member of the specified guild, and populates `AuthContext.GuildID`.
- [ ] If the user is not a member of the specified guild, the endpoint returns `403`.
- [ ] If `X-Guild-ID` is missing on a Bearer-auth request, the endpoint returns `400`.
- [ ] For Basic-auth requests, `X-Guild-ID` is ignored — the guild is derived from `api_clients.scope` as before.
- [ ] The guild membership check may use a short-TTL cache (e.g. 60 s) for performance.

### AC-6: User Attribution

- [ ] When a Bearer-authenticated user creates a grocery entry (`POST /groceries`), the `updated_by_id` field is set server-side to the Discord user ID from the JWT — regardless of what the client sends in the body.
- [ ] When a Bearer-authenticated user modifies a grocery entry (future edit endpoint), the same override applies.
- [ ] For Basic-auth requests, existing behavior is preserved (no server-side override of `updated_by_id`).

### AC-7: Backward Compatibility

- [ ] All existing API endpoints continue to work with `Authorization: Basic` (guild-scoped `api_clients` auth) exactly as before.
- [ ] The `/developer` slash command is unchanged.
- [ ] The `api_clients` table and its schema are unchanged.
- [ ] The `/metrics` endpoint remains unauthenticated.
- [ ] Existing CORS, rate limiting, timeout, and recovery middleware remain functional. CORS `AllowHeaders` includes `X-Guild-ID`.

### AC-8: Rate Limiting

- [ ] For Bearer-auth requests, the rate limiter keys on the Discord user ID (from the JWT `sub` claim) rather than the IP address.
- [ ] For Basic-auth requests, the existing rate limiter behavior (keyed on `clientID`) is preserved.

### AC-9: Database Migration

- [ ] A new migration file is added to `db/changelog/` that creates the `user_sessions` table.
- [ ] The migration runs automatically on startup, consistent with existing migration behavior.
- [ ] The migration is backward-compatible — it does not alter any existing tables.

### AC-10: Configuration

- [ ] OAuth2-related registration requires: **`DISCORD_CLIENT_ID`**, **`DISCORD_CLIENT_SECRET`**, **`SESSION_ENCRYPTION_KEY`**, and **`JWT_SIGNING_KEY`** (for Bearer on protected routes). **`ALLOWED_REDIRECT_URIS`** is optional (defaults to `grocerybot://auth/callback`) but must list every client `redirect_uri` used in production.
- [ ] If Discord OAuth / session encryption prerequisites are missing, **`/auth/*`** routes are not registered, but the rest of the API (including Basic auth) continues to work. The bot does not panic. (`/guilds` not registered until implemented.)
- [ ] Each `redirect_uri` the client sends in **`POST /auth/token`** must be allowlisted and registered in the Discord Developer Portal.

---

## 7. Decisions

All previously open questions have been resolved:

| # | Question | Decision |
|---|----------|----------|
| 1 | Authorization model | **Permissive.** Any guild member can CRUD groceries. No role/permission checks for MVP. (See §5.2.) |
| 2 | Guild selection UX | **`X-Guild-ID` header.** Bearer-auth requests pass the guild ID as a custom request header. Ignored for Basic auth. (See §5.6, AC-5.) |
| 3 | Token lifetimes | **15 min JWT, 7 day refresh token.** Refresh token rotation is optional. (See §4.4, AC-2, AC-3.) |
| 4 | Deployment | **In-process goroutine.** API stays in the bot process. All session/token state is DB-persisted (SQLite) — the bot may restart for updates. Guild resolution uses `discordgo.Session.State.Guilds` (rebuilt on startup). (See §5.1.) |
| 5 | Mobile platform | **React Native + Expo.** OAuth2 redirect URI uses a custom scheme or Expo AuthSession proxy. (See §5.11.) |
| 6 | Web app + Basic auth | **No.** The web/mobile app uses Bearer auth (JWT) exclusively. Basic auth is for existing server-admin integrations only; the app does not fall back to it. |
