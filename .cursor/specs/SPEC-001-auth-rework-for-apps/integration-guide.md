# GroceryBot API — Auth Integration Guide

This guide explains how an external app (e.g. GroceryBot App, built with React Native + Expo) authenticates users via Discord and makes guild-scoped API requests against the GroceryBot backend.

## Overview

The auth flow is **Discord OAuth2 Authorization Code + PKCE**, with the client driving the OAuth flow and the backend performing the confidential code exchange. The backend issues its own short-lived JWTs for API access and long-lived refresh tokens for session continuity.

```
App                          GroceryBot API              Discord
 │                                │                         │
 │ 1. Open Discord authorize URL  │                         │
 │   (with PKCE + state)          │                         │
 │───────────────────────────────────────────────────────────►
 │                                │                         │
 │ 2. User approves, Discord      │                         │
 │   redirects to app with code   │                         │
 │◄──────────────────────────────────────────────────────────│
 │                                │                         │
 │ 3. POST /auth/token            │                         │
 │   { code, code_verifier,       │                         │
 │     redirect_uri }             │                         │
 │───────────────────────────────►│                         │
 │                                │ 4. Exchange code+secret │
 │                                │───────────────────────►│
 │                                │◄───────────────────────│
 │                                │ 5. GET /users/@me      │
 │                                │───────────────────────►│
 │                                │◄───────────────────────│
 │                                │                         │
 │ 6. { access_token,             │                         │
 │      refresh_token,            │                         │
 │      expires_in }              │                         │
 │◄───────────────────────────────│                         │
 │                                │                         │
 │ 7. GET /guilds                 │                         │
 │   Authorization: Bearer <jwt>  │                         │
 │───────────────────────────────►│                         │
 │                                │ (fetches user guilds   │
 │                                │  from Discord, filters │
 │                                │  to bot-installed ones)│
 │ 8. [{ id, name, icon }, ...]   │                         │
 │◄───────────────────────────────│                         │
 │                                │                         │
 │ 9. CRUD (e.g. GET /grocery-    │                         │
 │   lists)                       │                         │
 │   Authorization: Bearer <jwt>  │                         │
 │   X-Guild-ID: <guild_id>      │                         │
 │───────────────────────────────►│                         │
 │ 10. Response                   │                         │
 │◄───────────────────────────────│                         │
```

---

## Prerequisites

### Discord Developer Portal

1. Go to https://discord.com/developers/applications and select the GroceryBot application.
2. Under **OAuth2 > Redirects**, add the redirect URI(s) the app will use. Examples:
   - `grocerybot://auth/callback` (custom scheme for React Native)
   - Expo AuthSession proxy URL, if used during development
3. Note the **Client ID**. The client secret stays server-side only — the app never needs it.

### Backend Configuration

The backend operator must set these env vars for the auth routes to be registered:

| Variable | Purpose |
|----------|---------|
| `DISCORD_CLIENT_ID` | Discord application client ID |
| `DISCORD_CLIENT_SECRET` | Used server-side to exchange authorization codes |
| `SESSION_ENCRYPTION_KEY` | Hex-encoded 32-byte AES key for encrypting Discord tokens at rest |
| `JWT_SIGNING_KEY` | HMAC key for signing GroceryBot JWTs |
| `ALLOWED_REDIRECT_URIS` | Comma-separated allowlist of redirect URIs the app may use (default: `grocerybot://auth/callback`) |

If any of `DISCORD_CLIENT_ID`, `DISCORD_CLIENT_SECRET`, or `SESSION_ENCRYPTION_KEY` are missing, the `/auth/*` routes are not registered. The rest of the API (including Basic auth) continues to work.

---

## 1. Login — Discord OAuth2 + PKCE

The app drives the entire OAuth2 flow client-side. The backend never redirects the user to Discord; it only receives the authorization code after the fact.

### 1.1 Build the Authorization URL

Construct the Discord authorization URL with these parameters:

| Parameter | Value |
|-----------|-------|
| `response_type` | `code` |
| `client_id` | The Discord application client ID (same as `DISCORD_CLIENT_ID` on the backend) |
| `redirect_uri` | One of the URIs registered in the Discord Developer Portal and listed in `ALLOWED_REDIRECT_URIS` (e.g. `grocerybot://auth/callback`) |
| `scope` | `identify guilds` |
| `state` | A random string generated by the client to prevent CSRF. Verify it when the redirect comes back. |
| `code_challenge` | The S256 PKCE challenge derived from a random `code_verifier` |
| `code_challenge_method` | `S256` |

Base URL: `https://discord.com/api/oauth2/authorize`

### 1.2 Open the System Browser

Open the constructed URL in the system browser (not an embedded webview). On Expo, use `expo-auth-session` / `expo-web-browser` which handle this automatically.

### 1.3 Handle the Redirect

Discord redirects the user back to your `redirect_uri` with `?code=<authorization_code>&state=<state>`. The app should:

1. Verify `state` matches what was generated in step 1.1.
2. Extract the `code` parameter.

### 1.4 Exchange the Code

Send the authorization code to the GroceryBot backend:

```
POST /auth/token
Content-Type: application/json

{
  "code": "<authorization_code>",
  "code_verifier": "<pkce_code_verifier>",
  "redirect_uri": "grocerybot://auth/callback"
}
```

All three fields are required. The `redirect_uri` must exactly match the one used in the authorization URL and must be in the backend's `ALLOWED_REDIRECT_URIS`.

**Success response (200):**

```json
{
  "access_token": "<grocerybot_jwt>",
  "refresh_token": "<opaque_refresh_token>",
  "expires_in": 900
}
```

| Field | Description |
|-------|-------------|
| `access_token` | A GroceryBot JWT (not a Discord token). Use this in `Authorization: Bearer` headers. |
| `refresh_token` | An opaque token for obtaining new access tokens without re-authenticating. |
| `expires_in` | Access token lifetime in seconds (currently 900 = 15 minutes). |

**Error responses:**

| Status | Meaning |
|--------|---------|
| `400` | Missing fields, or `redirect_uri` is not allowlisted |
| `502` | Discord code exchange or identity verification failed |
| `500` | Internal error (JWT issuer not ready, encryption failure, etc.) |

### 1.5 Store Tokens

Store `access_token` and `refresh_token` in secure storage (e.g. `expo-secure-store` on React Native). The access token is short-lived; the refresh token is long-lived (7 days).

---

## 2. Making Authenticated Requests

### 2.1 Guild-Scoped Endpoints (CRUD)

All CRUD endpoints require two things for Bearer-authenticated requests:

1. **`Authorization: Bearer <access_token>`** header
2. **`X-Guild-ID: <guild_id>`** header — specifies which Discord guild the request is scoped to

```
GET /grocery-lists
Authorization: Bearer <jwt>
X-Guild-ID: 123456789012345678
```

The backend verifies that:
- The JWT is valid and not expired.
- The user (from JWT `sub` claim) is a member of the specified guild.
- GroceryBot is installed in that guild.

If `X-Guild-ID` is missing, the backend returns **400**. If the user is not a member or the bot is not in the guild, it returns **403**.

### 2.2 Endpoints That Don't Require X-Guild-ID

These endpoints require `Authorization: Bearer` but **not** `X-Guild-ID`:

| Endpoint | Purpose |
|----------|---------|
| `GET /guilds` | List guilds the user can interact with (not yet implemented) |
| `POST /auth/logout` | End the user's session |

### 2.3 Endpoints That Don't Require Auth

| Endpoint | Purpose |
|----------|---------|
| `POST /auth/token` | Exchange Discord authorization code for tokens |
| `POST /auth/refresh` | Refresh an expired access token |
| `GET /metrics` | Prometheus metrics |

---

## 3. Available API Endpoints

### Auth

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/auth/token` | None | Exchange Discord auth code for GroceryBot tokens |
| `POST` | `/auth/refresh` | None | Refresh an expired access token |
| `POST` | `/auth/logout` | Bearer | End session (deletes server-side session data) |

### Guilds

| Method | Path | Auth | X-Guild-ID | Description |
|--------|------|------|------------|-------------|
| `GET` | `/guilds` | Bearer | No | List guilds the user can interact with *(not yet implemented)* |

### Grocery CRUD

| Method | Path | Auth | X-Guild-ID | Description |
|--------|------|------|------------|-------------|
| `GET` | `/grocery-lists` | Bearer | **Required** | List grocery entries and lists for the guild |
| `POST` | `/groceries` | Bearer | **Required** | Create a grocery entry |
| `DELETE` | `/groceries/:id` | Bearer | **Required** | Delete a grocery entry |

### Other

| Method | Path | Auth | X-Guild-ID | Description |
|--------|------|------|------------|-------------|
| `GET` | `/registrations` | Bearer | **Required** | List guild registrations/tier info |
| `GET` | `/metrics` | None | No | Prometheus metrics |

---

## 4. Token Refresh

Access tokens expire after **15 minutes** (`expires_in: 900`). The app should refresh proactively (e.g. when `expires_in` is close to 0, or on receiving a 401).

```
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "<opaque_refresh_token>"
}
```

**Success response (200):**

```json
{
  "access_token": "<new_jwt>",
  "refresh_token": "<new_refresh_token>",
  "expires_in": 900
}
```

The refresh token is **rotated** on every use — the old one is invalidated and a new one is returned. Always store the new `refresh_token` from the response.

**Error responses:**

| Status | Meaning | App Action |
|--------|---------|------------|
| `401` | Refresh token expired or invalid | Redirect user to login (full Discord OAuth flow) |
| `400` | Missing refresh token in body | Fix the request |

Refresh tokens are valid for **7 days**. If the user doesn't open the app for 7 days, they'll need to log in again via Discord.

---

## 5. Guild Selection

After login, the app should let the user pick which Discord guild to interact with.

### 5.1 Fetching the Guild List

> **Note:** `GET /guilds` is not yet implemented on the backend. This section describes the planned interface.

```
GET /guilds
Authorization: Bearer <jwt>
```

**Response (200):**

```json
{
  "guilds": [
    { "id": "123456789012345678", "name": "Flatmate's Server", "icon": "abc123" },
    { "id": "987654321098765432", "name": "My Server", "icon": "def456" }
  ]
}
```

The list is the **intersection** of:
- Guilds the user belongs to (fetched live from Discord using the user's stored Discord token)
- Guilds where GroceryBot is installed

### 5.2 Using the Selected Guild

Once the user selects a guild, include its `id` as the `X-Guild-ID` header on all subsequent CRUD requests.

If the stored Discord token has expired, the backend will attempt to refresh it automatically. If it can't be refreshed (e.g. user revoked access), the endpoint returns **401** — the app should prompt the user to log in again.

---

## 6. Logout

```
POST /auth/logout
Authorization: Bearer <jwt>
```

**Response:** `204 No Content`

This deletes all server-side session data for the user (including stored Discord tokens). The app should clear its locally stored `access_token` and `refresh_token`.

No request body or `X-Guild-ID` header is needed.

---

## 7. Expo / React Native Integration Notes

### Recommended Libraries

| Library | Purpose |
|---------|---------|
| `expo-auth-session` | Manages PKCE, state, authorization URL construction, and redirect parsing |
| `expo-web-browser` | Opens the system browser for the OAuth flow |
| `expo-secure-store` | Securely stores JWT and refresh token on-device |
| `expo-linking` | Handles deep link / custom scheme registration |

### Example Flow (Pseudocode)

```typescript
import * as AuthSession from 'expo-auth-session';
import * as WebBrowser from 'expo-web-browser';

const discovery = {
  authorizationEndpoint: 'https://discord.com/api/oauth2/authorize',
};

const DISCORD_CLIENT_ID = '<your_client_id>';
const REDIRECT_URI = 'grocerybot://auth/callback';
const API_BASE = 'https://api.grocerybot.net'; // or your backend URL

// 1. Create the auth request (PKCE is enabled by default)
const [request, response, promptAsync] = AuthSession.useAuthRequest(
  {
    clientId: DISCORD_CLIENT_ID,
    redirectUri: REDIRECT_URI,
    scopes: ['identify', 'guilds'],
    responseType: AuthSession.ResponseType.Code,
  },
  discovery,
);

// 2. Open the browser
await promptAsync();

// 3. Handle the response
if (response?.type === 'success') {
  const { code } = response.params;

  // 4. Exchange with GroceryBot backend
  const tokenResponse = await fetch(`${API_BASE}/auth/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      code,
      code_verifier: request.codeVerifier,
      redirect_uri: REDIRECT_URI,
    }),
  });

  const { access_token, refresh_token, expires_in } = await tokenResponse.json();

  // 5. Store tokens securely
  await SecureStore.setItemAsync('access_token', access_token);
  await SecureStore.setItemAsync('refresh_token', refresh_token);
}
```

### Making an Authenticated Request

```typescript
const response = await fetch(`${API_BASE}/grocery-lists`, {
  headers: {
    'Authorization': `Bearer ${accessToken}`,
    'X-Guild-ID': selectedGuildId,
  },
});
```

### Refreshing Tokens

```typescript
async function refreshAccessToken(currentRefreshToken: string) {
  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: currentRefreshToken }),
  });

  if (res.status === 401) {
    // Refresh token expired — redirect to login
    return null;
  }

  const { access_token, refresh_token, expires_in } = await res.json();
  // Store both new tokens
  await SecureStore.setItemAsync('access_token', access_token);
  await SecureStore.setItemAsync('refresh_token', refresh_token);
  return { access_token, refresh_token, expires_in };
}
```

---

## 8. Rate Limiting

The API rate-limits at **10 requests per 30 seconds**, keyed by Discord user ID for Bearer-authenticated requests. If you exceed this, the API returns **429 Too Many Requests**.

---

## 9. CORS

Native React Native requests (not from a webview) don't send an `Origin` header, so CORS does not apply. If the app has a web build, the backend's `GROCER_BOT_API_ALLOW_ORIGINS` must include the web app's origin. The `X-Guild-ID` header is already in the CORS `AllowHeaders` list.

---

## 10. Error Handling Summary

| Status | Meaning | Recommended Action |
|--------|---------|-------------------|
| `400` | Bad request (missing fields, missing `X-Guild-ID`, invalid `redirect_uri`) | Fix the request |
| `401` | Token expired, invalid, or refresh token revoked | Refresh the access token; if refresh fails, redirect to login |
| `403` | User not a member of the specified guild, or bot not in guild | Show an error; prompt guild re-selection |
| `429` | Rate limited | Back off and retry |
| `500` | Server error | Retry or show error |
| `502` | Discord API unreachable or returned an error | Retry later |
