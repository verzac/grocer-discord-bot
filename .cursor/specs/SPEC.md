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

### 2.5 Assumed Environment Variables for OAuth2

The following env vars are assumed to be pre-configured by the operator. They are **not currently used** anywhere in the codebase (the hardcoded client ID `815120759680532510` in invite URLs is a different concern — it's the bot's application ID for the OAuth2 bot-invite flow, not the user-facing OAuth2 flow).

| Variable | Purpose |
|----------|---------|
| `DISCORD_CLIENT_ID` | Discord application client ID for the OAuth2 Authorization Code flow. |
| `DISCORD_CLIENT_SECRET` | Discord application client secret for exchanging authorization codes for access tokens (server-side only). |
| `DISCORD_REDIRECT_URI` | The URL that Discord redirects the user back to after they authorize. See explanation below. |

These come from the Discord Developer Portal under the same application that owns the bot.

**`DISCORD_REDIRECT_URI` explained:**

OAuth2 Authorization Code flow works in two hops:

1. The client sends the user to Discord's authorization page: `https://discord.com/oauth2/authorize?client_id=...&redirect_uri=...&response_type=code&scope=identify+guilds`
2. After the user approves, Discord redirects their browser to the `redirect_uri` with an authorization `code` appended as a query parameter: `<redirect_uri>?code=abc123`

`DISCORD_REDIRECT_URI` is that callback URL — the GroceryBot endpoint that receives the authorization code and exchanges it for tokens. For example:

- Production: `https://api.grocerybot.net/auth/discord/callback`
- Local dev: `http://localhost:8080/auth/discord/callback`

This value must **exactly match** one of the redirect URIs registered in the Discord Developer Portal (under the application's OAuth2 settings). Discord will reject the authorization request if the `redirect_uri` parameter doesn't match a registered URI. This is a security measure to prevent authorization codes from being sent to attacker-controlled URLs.

The env var is needed (rather than hardcoding) because the value differs across environments (local dev, staging, production).

Additional env vars introduced by this approach (e.g. `JWT_SIGNING_KEY`) are listed in §4.2.

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

The GroceryBot backend drives the full OAuth2 Authorization Code flow with Discord and issues its own short-lived JWTs.

### 4.1 Flow

1. **Client** opens `https://api.grocerybot.net/auth/discord` (or a configured URL) which redirects to Discord's authorization page.
2. **Discord** redirects back to `https://api.grocerybot.net/auth/discord/callback?code=...`.
3. **GroceryBot backend** exchanges the `code` for a Discord access token + refresh token (server-side, confidential).
4. Backend calls Discord `GET /users/@me` to get the user's identity.
5. Backend stores the user's Discord access token (encrypted) and refresh token server-side, associated with the user ID. This is used later for guild lookups.
6. Backend issues a **GroceryBot JWT** (short-lived access token) containing `{ sub: discordUserID, exp, iat }` and a **refresh token** stored server-side. The JWT identifies the user only — **it does not contain a guild list**. Refresh tokens are required so users don't have to re-authenticate with Discord every time the access token expires.
7. Client stores the JWT (in-memory or secure storage) and includes it as `Authorization: Bearer <jwt>` on subsequent requests.
8. When the user wants to select a guild, `GET /guilds` calls Discord `GET /users/@me/guilds` using the **stored Discord access token** for that user, cross-references with bot-installed guilds, and returns a **fresh** list. This means the guild list is always up-to-date at the moment of selection.
9. CRUD endpoints accept a `guild_id` query parameter (or path segment). The middleware verifies the user's identity from the JWT; the handler verifies that the user is actually a member of the requested guild (either by a fresh Discord call or a short-lived cache).

### 4.2 New/Changed Components

| Component | Change |
|-----------|--------|
| **DB** | New `user_sessions` table to store Discord access/refresh tokens (encrypted) keyed by user ID, plus GroceryBot refresh tokens. |
| **Routes** | `GET /auth/discord` — initiate OAuth2 redirect. |
|  | `GET /auth/discord/callback` — exchange code, issue JWT. |
|  | `POST /auth/refresh` — exchange a refresh token for a new JWT access token. |
|  | `GET /guilds` — fetch the user's guilds **live from Discord** (using stored Discord token), intersect with bot guilds, return fresh list. |
| **Middleware** | Extend `AuthMiddleware` to accept both `Basic` and `Bearer` schemes. Bearer path verifies JWT signature + expiry, extracts user ID. |
| **Handlers** | Existing CRUD handlers resolve `guildID` from a request param instead of (only) from scope. Handler/middleware verifies user's guild membership (via cached Discord lookup or short-TTL cache). |
| **Config** | New env var: `JWT_SIGNING_KEY`. Uses `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, and `DISCORD_REDIRECT_URI` from §2.5. |
| **Dependencies** | A JWT library (e.g. `golang-jwt/jwt`). |

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

1. On login, the backend issues both a **short-lived JWT** (e.g. 15 min) and a **long-lived refresh token** (e.g. 7–30 days) stored in the `user_sessions` table.
2. The client stores both. When the JWT expires, the client calls `POST /auth/refresh` with the refresh token.
3. The backend validates the refresh token (lookup by hash in `user_sessions`, check expiry), then issues a new JWT. Optionally rotates the refresh token (issue a new one, invalidate the old).
4. If the refresh token itself is expired or revoked, the user must re-authenticate with Discord.

This means a user stays logged in for the refresh token's lifetime without ever seeing Discord's OAuth2 page again.

### 4.5 Implementation: Auth0 vs Self-Rolled vs Go Libraries

There are three implementation paths for the OAuth2 + JWT layer. Below is an evaluation.

#### Path 1 — Auth0 (Managed Identity Provider)

Auth0 supports Discord as a social connection out of the box. It handles the OAuth2 flow, token issuance, refresh tokens, and user management.

**How it would work:**

- Configure Discord as a social connection in Auth0 Dashboard.
- The client app uses Auth0's SDK to log in (Auth0 redirects to Discord, handles callback, issues Auth0 JWTs + refresh tokens).
- GroceryBot API validates Auth0 JWTs using Auth0's JWKS endpoint or a shared signing key.
- To call Discord's `GET /users/@me/guilds`, we need the user's **Discord access token** — Auth0 can store this via its **Token Vault / Connected Accounts** feature.

**Pricing:** Free tier covers 25,000 MAU, unlimited social connections, no credit card required. More than sufficient for GroceryBot's scale.

**Verdict: Overkill for this use case.** Here's why:

| Concern | Detail |
|---------|--------|
| **Guild access** | We need to call Discord's API with the user's Discord token to fetch guilds. Auth0 doesn't natively expose the upstream provider's access token to your API — you need Token Vault (Connected Accounts) to retrieve it, which adds complexity and an Auth0 Management API call on the hot path. |
| **Indirection** | Auth0 sits between us and Discord. We'd authenticate with Auth0, then separately fetch Discord tokens from Auth0's Token Vault to call Discord APIs. Two hops where one suffices. |
| **Operational dependency** | Adds Auth0 as a runtime dependency. If Auth0 is down, users can't log in or refresh tokens — even though Discord and our DB are fine. |
| **Scope mismatch** | Auth0 is designed for apps with multiple identity providers (Google, GitHub, email/password, SSO). GroceryBot has exactly one: Discord. The multi-provider abstraction adds overhead we don't need. |
| **Custom scopes** | We need `identify` + `guilds` scopes from Discord. Configurable in Auth0, but requires careful setup via `connection_scope` — not the default path. |
| **Refresh tokens** | Auth0 handles refresh tokens for its own JWTs (good), but we also need to refresh the **Discord** token stored in Token Vault for guild lookups. This means managing two refresh cycles. |

Auth0 would make sense if GroceryBot needed multi-provider auth (e.g. Google + Discord + email/password), enterprise SSO, or MFA. For a single Discord-only login with guild-specific API access, it introduces more complexity than it removes.

#### Path 2 — Self-Rolled (Manual OAuth2 + JWT)

Write the OAuth2 code exchange, JWT issuance, and refresh token logic directly using `golang.org/x/oauth2` + `golang-jwt/jwt/v5`.

**Pros:**
- Minimal dependencies; full control.
- No abstraction mismatch — we call Discord directly.
- `golang-jwt/jwt/v5` (v5.3.1) is the standard Go JWT library, well-maintained, supports HMAC/RSA/ECDSA signing.
- `golang.org/x/oauth2` is Go's standard OAuth2 client — handles code exchange, token refresh, and HTTP client injection.

**Cons:**
- Must implement: state parameter generation/validation, PKCE, callback handler, token storage, refresh logic, JWT signing/verification middleware.
- More boilerplate than using a framework, but the total surface area is small (a few hundred lines).

#### Path 3 — Go OAuth2/OIDC Libraries with Discord Support

Several Go libraries wrap the OAuth2 flow with Discord-specific providers:

| Library | What it does | Refresh tokens | Discord guilds | Notes |
|---------|-------------|:-:|:-:|-------|
| **`markbates/goth`** (v1.82) | Multi-provider OAuth2 framework with session management. Discord provider built-in. | Yes (`RefreshToken()`) | Not built-in (call Discord API yourself) | Most established; 50+ providers. Brings its own session abstraction (cookie/gorilla). May conflict with Echo's patterns. |
| **`disgoorg/disgo/oauth2`** | Discord-specific OAuth2 module with state management. | Yes | Yes (Discord-native SDK) | Part of the larger `disgo` Discord library. More Discord-aware than generic OAuth2 libs, but pulls in a larger dependency tree. |
| **`notclynt/discord-oauth2`** | Thin Discord endpoint config for `golang.org/x/oauth2`. | Via `golang.org/x/oauth2` | Not built-in | Extremely lightweight — just provides the Discord `oauth2.Endpoint`. Almost identical to self-rolling with `golang.org/x/oauth2`. |
| **`zekith/oauthgo`** (2025) | Modern OAuth2/OIDC lib with 50+ providers, PKCE, state/nonce validation. | Yes | Not built-in | Newer, less battle-tested. Good security defaults (PKCE, replay protection). |

#### Recommendation

**Path 2 (self-rolled) or `notclynt/discord-oauth2`** (which is effectively Path 2 with pre-configured Discord endpoints). The reasoning:

1. **Discord is not an OIDC provider** — it implements OAuth2 but not the OpenID Connect layer (no `id_token`, no `.well-known/openid-configuration`). OIDC-focused libs don't add value here.
2. **We need guild access** — no library handles the `GET /users/@me/guilds` → intersect with bot guilds → cache pattern. We'll write that regardless.
3. **JWT issuance is ours** — we issue GroceryBot JWTs, not Discord tokens. The JWT layer is independent of the OAuth2 library.
4. **Refresh is straightforward** — `golang.org/x/oauth2` already handles Discord token refresh. Our own refresh tokens are a simple DB lookup + new JWT.
5. **Minimal coupling** — `goth` and `disgo` bring their own session/state abstractions that may conflict with Echo middleware patterns. The lighter the dependency, the easier to maintain.

**Concrete dependencies:**

| Package | Purpose |
|---------|---------|
| `golang.org/x/oauth2` | OAuth2 code exchange, Discord token refresh. |
| `github.com/golang-jwt/jwt/v5` | Sign and verify GroceryBot JWTs. |

Optionally add `github.com/notclynt/discord-oauth2` for pre-configured Discord endpoints (saves ~5 lines of config), but it's not required.

---

## 5. Shared Concerns

### 5.1 Guild Resolution

The `GET /guilds` endpoint and CRUD guild-membership checks need a way to determine which guilds the bot is installed in. Two approaches:

- **A) Bot session state** — `discordgo.Session.State.Guilds` holds the bot's current guild list in memory. Fast, but only available in the bot process, not a separate API process.
- **B) DB-derived** — query distinct `guild_id` values from `guild_configs` or `grocery_entries`. Slightly stale if the bot joins a guild but no config row exists yet.

Since the API runs in-process (goroutine in `main.go`), approach **(A)** is available and preferred.

### 5.2 Authorization Within a Guild

Currently, any holder of the guild API key can perform any operation. For user-scoped auth, decide:

- **Permissive (recommended for MVP):** Any guild member can CRUD groceries on that guild. This mirrors the bot's Discord command behavior (no role restrictions except `/developer`).
- **Role-based (future):** Check the user's Discord roles/permissions in the target guild before allowing mutations. Requires additional Discord API calls or caching role info.

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
    UserID  string // Bearer auth: Discord user ID (new)
    GuildID string // resolved from scope or request param (new)
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

Currently, the guild is implicitly derived from the API key's scope. With user-scoped auth, the guild must be **explicitly specified** per request. Two patterns:

- **Query parameter:** `GET /grocery-lists?guild_id=123456789`
- **Path prefix:** `GET /guilds/123456789/grocery-lists`

The path-prefix approach is more RESTful and avoids the risk of forgetting the query parameter, but either works. The middleware must verify the user has access to the specified guild.

### 5.7 CORS

The allowed origins list (`GROCER_BOT_API_ALLOW_ORIGINS`) will need to include the mobile app's web view origin (if applicable) or the web app's domain.

### 5.8 Rate Limiting

The current rate limiter keys on `clientID` (Basic auth) or IP. For Bearer auth, it should key on the Discord user ID to prevent per-user abuse while allowing shared IPs (e.g. NAT).

### 5.9 Discord OAuth2 Scopes Needed

The Discord OAuth2 request needs these scopes:

- `identify` — access to `GET /users/@me` (user ID, username, avatar).
- `guilds` — access to `GET /users/@me/guilds` (list of guilds the user is in).

No bot-level or message-level scopes are required.

### 5.10 PKCE

For public clients (SPA, mobile app), **PKCE (Proof Key for Code Exchange)** should be used with the Authorization Code flow. Discord supports PKCE. This removes the need for the client to hold a client secret.

### 5.11 Missing CRUD Endpoints

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

## 6. Open Questions

1. **Authorization model:** Should any guild member be able to CRUD, or should we restrict to certain Discord roles/permissions?
2. **Guild selection UX:** Query parameter (`?guild_id=`) or path prefix (`/guilds/:guildID/...`) for guild-scoped endpoints?
3. **Token lifetimes:** What TTLs for the JWT access token and refresh token? (Suggestion: 15 min access / 7 day refresh. Refresh token rotation optional.)
4. **Deployment:** Does the API remain an in-process goroutine, or will it be separated into its own service in the future? This affects guild resolution (bot session state vs DB).
5. **Mobile platform:** Native iOS/Android or a web-based app (React Native, Flutter, PWA)? This affects the OAuth2 redirect URI scheme (`https://` vs custom scheme like `grocerybot://`).
6. **Do we need to support the web app using the existing Basic auth as well?** (e.g. should the web app be able to fall back to a guild API key if the user is not logged in via Discord?)
