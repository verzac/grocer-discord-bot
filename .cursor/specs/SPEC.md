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

## 4. Options

Three options are presented below. They differ primarily in **where tokens are issued** and **how the API resolves the user's identity on each request**.

---

### Option A — Discord OAuth2 + GroceryBot-Issued JWT

The GroceryBot backend drives the full OAuth2 Authorization Code flow with Discord and issues its own short-lived JWTs.

#### Flow

1. **Client** opens `https://api.grocerybot.net/auth/discord` (or a configured URL) which redirects to Discord's authorization page.
2. **Discord** redirects back to `https://api.grocerybot.net/auth/discord/callback?code=...`.
3. **GroceryBot backend** exchanges the `code` for a Discord access token + refresh token (server-side, confidential).
4. Backend calls Discord `GET /users/@me` and `GET /users/@me/guilds` using the Discord access token.
5. Backend cross-references the user's guild list against guilds where the bot is installed (the bot's own `discordgo.Session.State.Guilds` or a DB lookup on known guild IDs).
6. Backend issues a **GroceryBot JWT** containing `{ sub: discordUserID, guilds: [allowedGuildIDs], exp, iat }` and optionally a **refresh token** stored server-side.
7. Client stores the JWT (in-memory or secure storage) and includes it as `Authorization: Bearer <jwt>` on subsequent requests.
8. A new `/guilds` endpoint returns the user's eligible guilds (derived from JWT claims or re-fetched).
9. CRUD endpoints accept a `guild_id` query parameter (or path segment) and the middleware verifies the guild is in the JWT's allowed list.

#### New/Changed Components

| Component | Change |
|-----------|--------|
| **DB** | New `user_sessions` table (optional — only if using opaque refresh tokens). |
| **Routes** | `GET /auth/discord` — initiate OAuth2 redirect. |
|  | `GET /auth/discord/callback` — exchange code, issue JWT. |
|  | `POST /auth/refresh` — refresh JWT (if refresh tokens are used). |
|  | `GET /guilds` — list eligible guilds for the authenticated user. |
| **Middleware** | Extend `AuthMiddleware` to accept both `Basic` and `Bearer` schemes. Bearer path verifies JWT signature + expiry, extracts user ID and guild list. |
| **Handlers** | Existing CRUD handlers resolve `guildID` from a request param instead of (only) from scope. Middleware verifies the guild is in the JWT's claims. |
| **Config** | New env vars: `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, `DISCORD_REDIRECT_URI`, `JWT_SIGNING_KEY`. |
| **Dependencies** | A JWT library (e.g. `golang-jwt/jwt`). |

#### Pros

- **Full control** over token lifetime, claims, and revocation.
- **No per-request Discord API call** — JWT is self-contained and verified locally.
- **Works for both confidential and public clients** — the OAuth2 code exchange happens server-side so the client never sees the Discord client secret.
- Supports **PKCE** extension for mobile (public) clients.
- Can embed extra claims (user roles, tier info) to avoid DB lookups on hot paths.

#### Cons

- More backend implementation work (OAuth2 callback handler, JWT issuance, refresh logic).
- Guild list in the JWT can go **stale** if the user joins/leaves guilds mid-session. Mitigation: short JWT expiry (e.g. 15 min) with refresh, or re-validate guild membership on sensitive operations.
- Requires secure storage of `JWT_SIGNING_KEY` and `DISCORD_CLIENT_SECRET`.

---

### Option B — Discord OAuth2 Passthrough (Use Discord Tokens Directly)

The client app handles the Discord OAuth2 flow itself (e.g. using Discord's SDK). The GroceryBot API receives the Discord access token on every request and validates it upstream.

#### Flow

1. **Client** performs Discord OAuth2 (Authorization Code + PKCE for public clients) directly with Discord, obtaining an access token with scopes `identify guilds`.
2. Client sends `Authorization: Bearer <discordAccessToken>` to GroceryBot API.
3. **GroceryBot middleware** calls Discord `GET /users/@me` (and optionally `GET /users/@me/guilds`) to validate the token and extract the user's identity + guild memberships.
4. Middleware caches the validation result briefly (e.g. 60 s in-memory) keyed by token hash to avoid hammering Discord.
5. CRUD endpoints accept `guild_id` as a request parameter; middleware checks it against the user's guild list intersected with bot-installed guilds.

#### New/Changed Components

| Component | Change |
|-----------|--------|
| **DB** | None required (stateless). |
| **Routes** | `GET /guilds` — returns eligible guilds (fetched from Discord and filtered). |
| **Middleware** | New `BearerDiscordMiddleware` that validates Discord tokens upstream. Dual-mode: Basic for legacy, Bearer for new flow. |
| **Handlers** | Same CRUD changes as Option A (guild from request param). |
| **Config** | Potentially `DISCORD_CLIENT_ID` and `DISCORD_CLIENT_SECRET` only if the backend needs to refresh tokens on behalf of the user. Otherwise, none. |
| **Cache** | Extend `cache/` package for token-validation caching. |

#### Pros

- **Simplest backend** — no token issuance, no JWT signing, no refresh logic.
- **No new DB tables** for sessions.
- **Always-fresh** guild list (fetched from Discord on each request or short TTL cache).

#### Cons

- **Per-request Discord API call** (or per cache-miss). Discord API rate limits (50 req/s per token globally) may become a bottleneck under moderate load.
- **Latency** — each uncached request adds ~100–300 ms for the Discord roundtrip.
- **Token management on the client** — the client must handle Discord token refresh itself, which adds complexity to mobile/web apps.
- **Less control** — cannot add custom claims; role/tier lookups always hit the DB.
- If Discord has an **outage**, the GroceryBot API becomes unusable even though the data layer is fine.

---

### Option C — Discord OAuth2 + GroceryBot-Issued Opaque Session Token

Similar to Option A, but instead of a JWT, the backend issues an **opaque session token** stored in a server-side session table.

#### Flow

1–5. Same as Option A (backend-driven OAuth2, Discord token exchange, guild resolution).
6. Backend generates a random opaque token (e.g. 64-char hex), stores it in a `user_sessions` table with `user_id`, `discord_access_token` (encrypted), `guild_ids`, `expires_at`.
7. Client receives the opaque token and sends it as `Authorization: Bearer <sessionToken>` on subsequent requests.
8. Middleware looks up the session in DB, verifies expiry, extracts user ID and guilds.
9. A background job or on-demand refresh re-validates guild membership via the stored Discord token.

#### New/Changed Components

| Component | Change |
|-----------|--------|
| **DB** | New `user_sessions` table: `token_hash`, `user_id`, `discord_user_id`, `discord_access_token` (encrypted), `discord_refresh_token` (encrypted), `guild_ids` (JSON), `expires_at`, `created_at`. |
| **Routes** | Same as Option A (`/auth/discord`, `/auth/discord/callback`, `/auth/refresh`, `/guilds`). |
| **Middleware** | Bearer path does a DB lookup by token hash instead of JWT verification. |
| **Config** | `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, `DISCORD_REDIRECT_URI`, `SESSION_ENCRYPTION_KEY`. |

#### Pros

- **Immediate revocation** — delete the session row and the token is invalid.
- **Can store Discord tokens server-side** for background guild refresh, avoiding stale data.
- **No JWT signing key** to manage or rotate.
- Familiar session-based pattern.

#### Cons

- **DB lookup on every request** — adds latency; mitigated by in-memory cache with short TTL.
- **More storage** — session table grows with active users; needs cleanup job.
- **Encrypted token storage** — storing Discord tokens at rest requires an encryption key and adds operational complexity.
- More implementation work than Option B.

---

## 5. Comparison Matrix

| Criteria | Option A (JWT) | Option B (Passthrough) | Option C (Opaque Session) |
|----------|:-:|:-:|:-:|
| Backend complexity | Medium | Low | Medium–High |
| Per-request external call | No | Yes (Discord) | No (DB lookup) |
| Latency overhead | Minimal | ~100–300 ms | Minimal (DB) |
| Token revocation | Hard (wait for expiry or blocklist) | N/A (Discord controls) | Instant (delete row) |
| Guild freshness | Stale until refresh | Always fresh | Stale until background refresh |
| Discord outage resilience | Unaffected (post-login) | API fully blocked | Unaffected (post-login) |
| Client complexity | Low (just store JWT) | Medium (handle Discord OAuth + refresh) | Low (just store token) |
| New DB tables | 0–1 | 0 | 1 |
| Dependency on Discord at runtime | Login only | Every request | Login + background refresh |

---

## 6. Shared Concerns (Regardless of Option)

### 6.1 Guild Resolution

All options need a way to determine which guilds the bot is installed in. Two approaches:

- **A) Bot session state** — `discordgo.Session.State.Guilds` holds the bot's current guild list in memory. Fast, but only available in the bot process, not a separate API process.
- **B) DB-derived** — query distinct `guild_id` values from `guild_configs` or `grocery_entries`. Slightly stale if the bot joins a guild but no config row exists yet.

Since the API runs in-process (goroutine in `main.go`), approach **(A)** is available and preferred.

### 6.2 Authorization Within a Guild

Currently, any holder of the guild API key can perform any operation. For user-scoped auth, decide:

- **Permissive (recommended for MVP):** Any guild member can CRUD groceries on that guild. This mirrors the bot's Discord command behavior (no role restrictions except `/developer`).
- **Role-based (future):** Check the user's Discord roles/permissions in the target guild before allowing mutations. Requires additional Discord API calls or caching role info.

### 6.3 Backward Compatibility

The existing Basic auth flow **must continue to work**. The middleware should support dual schemes:

```
Authorization: Basic ...   → existing api_clients lookup (guild-scoped)
Authorization: Bearer ...  → new user-scoped auth (JWT / session / Discord token)
```

No changes to the existing `api_clients` table or `/developer` slash command.

### 6.4 New Endpoint: `GET /guilds`

All options require a new endpoint that returns the guilds the authenticated user can interact with:

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

### 6.5 Guild Selection on CRUD Endpoints

Currently, the guild is implicitly derived from the API key's scope. With user-scoped auth, the guild must be **explicitly specified** per request. Two patterns:

- **Query parameter:** `GET /grocery-lists?guild_id=123456789`
- **Path prefix:** `GET /guilds/123456789/grocery-lists`

The path-prefix approach is more RESTful and avoids the risk of forgetting the query parameter, but either works. The middleware must verify the user has access to the specified guild.

### 6.6 CORS

The allowed origins list (`GROCER_BOT_API_ALLOW_ORIGINS`) will need to include the mobile app's web view origin (if applicable) or the web app's domain.

### 6.7 Rate Limiting

The current rate limiter keys on `clientID` (Basic auth) or IP. For Bearer auth, it should key on the Discord user ID to prevent per-user abuse while allowing shared IPs (e.g. NAT).

### 6.8 Discord OAuth2 Scopes Needed

Regardless of option, the Discord OAuth2 request needs these scopes:

- `identify` — access to `GET /users/@me` (user ID, username, avatar).
- `guilds` — access to `GET /users/@me/guilds` (list of guilds the user is in).

No bot-level or message-level scopes are required.

### 6.9 PKCE

For public clients (SPA, mobile app), **PKCE (Proof Key for Code Exchange)** should be used with the Authorization Code flow. Discord supports PKCE. This removes the need for the client to hold a client secret.

### 6.10 Missing CRUD Endpoints

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

## 7. Open Questions

1. **Which option do you prefer?** (A / B / C, or a hybrid.)
2. **Authorization model:** Should any guild member be able to CRUD, or should we restrict to certain Discord roles/permissions?
3. **Guild selection UX:** Query parameter (`?guild_id=`) or path prefix (`/guilds/:guildID/...`) for guild-scoped endpoints?
4. **Token lifetime:** For Option A/C, what should the access token TTL be? (Suggestion: 15 min access / 7 day refresh.)
5. **Deployment:** Does the API remain an in-process goroutine, or will it be separated into its own service in the future? This affects guild resolution (bot session state vs DB).
6. **Mobile platform:** Native iOS/Android or a web-based app (React Native, Flutter, PWA)? This affects the OAuth2 redirect URI scheme (`https://` vs custom scheme like `grocerybot://`).
7. **Do we need to support the web app using the existing Basic auth as well?** (e.g. should the web app be able to fall back to a guild API key if the user is not logged in via Discord?)
