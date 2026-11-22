# Intent tokens for Siri add-grocery

Author: verzac (github.com/verzac), drafted with Claude Code

## Why is this important?
1. The iOS app is adding a Siri App Intent ("Hey Siri, add to GroceryApp" → item). It runs in Swift, outside the React Native runtime, and must be able to `POST /groceries` while the phone is locked, without opening the app.
2. The normal session can't be shared with Swift:
   - The access JWT lasts 15 minutes (`auth.DefaultAccessTokenTTL`). A Siri invocation hours after the app was last opened will almost always hold an expired JWT.
   - `POST /auth/refresh` rotates the refresh token and the old one stops working. If Swift and JS both hold the refresh token, whichever refreshes second presents a revoked token and the user is logged out. Coordinating the two runtimes (shared keychain lock, single writer) is fragile and fails silently on a locked device.
3. Therefore, the app mints a dedicated **intent token** from JS (where the session lives) and hands it to Swift. Swift only presents it; it never refreshes anything.

## Design
- **Endpoint:** `POST /auth/intent-token`, authenticated with the normal Bearer JWT. No `X-Guild-ID`, no body. Returns `201 {"intent_token": "gbi_...", "expires_in": 2592000}`.
- **Scope:** the token is accepted only on `POST /groceries`. The auth middleware rejects it with `403` on every other method/path before looking it up. That includes `/auth/intent-token`, so an intent token cannot mint another one, and `/auth/logout`. A leaked token can add grocery entries to guilds the user is in. It cannot read, delete, or change anything.
- **Guild checks:** the request still sends `X-Guild-ID`. The same user-in-guild and bot-in-guild checks as the JWT path run (one shared helper), so revoking the user's guild membership or kicking the bot still blocks the token.
- **Lifetime:** fixed 30 days from minting, not extended on use. The app re-mints from JS while it has a live session; expiry bounds the damage if the device is lost without a logout.
- **One token per user:** minting deletes the user's existing intent tokens, then creates the new one.
- **Revocation:** `POST /auth/token` (new login) and `POST /auth/logout` delete the user's intent tokens. `POST /auth/refresh` does not touch them, since refresh happens constantly and must not break Siri. There is no `DELETE` endpoint; logout or re-mint covers it.
- **Storage:** only the SHA-256 hex hash of the full plaintext token is stored (`intent_tokens.token_hash`, unique index), the same scheme as refresh tokens. A database leak does not yield usable tokens. `last_used_at` is updated on each accepted request on a best-effort basis (a failed update is logged and does not fail the request).
- **Rate limiting:** the middleware sets the request identifier to the Discord user ID after the lookup, so intent token requests share the existing per-user rate limit with the user's JWT requests.

## Why the `gbi_` prefix?
1. **Routing:** both token types arrive as `Authorization: Bearer ...`. JWTs always start with `eyJ` (base64 of `{"`), so a `gbi_` prefix tells the middleware which verifier to use without trying JWT parsing first and without a new header or auth scheme.
2. **Log redaction:** a fixed prefix makes the token easy to match and mask in logs and crash reports.
3. **Secret scanning:** a distinctive prefix plus fixed-length base64url body (`gbi_[A-Za-z0-9_-]{43}`) can be registered as a custom secret-scanning pattern, so a token committed or pasted somewhere is detectable.

## Known limitations
- **Demo login:** `POST /auth/demo-login` reuses a single seeded session for the shared demo account and creates no session per device. All demo devices therefore share one intent token slot: each mint replaces the previous one, so only the most recently minted demo device's Siri token works. Demo logout returns early (as before) and does not revoke the demo intent token, so it doesn't log out other demo devices.
- **Concurrent mints:** delete-then-create is not wrapped in a transaction. Two simultaneous mints by the same user can leave two valid tokens until the next mint, login, or logout. Both are still scoped to the same user and endpoint.
