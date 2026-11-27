# Intent tokens for Siri add-grocery

Author: verzac (github.com/verzac), drafted with Claude Code

## Why is this important?
1. The iOS app is adding a Siri App Intent ("Hey Siri, add to GroceryApp" → item). It runs in Swift, outside the React Native runtime, and must be able to `POST /groceries` while the phone is locked, without opening the app.
2. The normal session can't be shared with Swift:
   - The access JWT lasts 15 minutes (`auth.DefaultAccessTokenTTL`). A Siri invocation hours after the app was last opened will almost always hold an expired JWT.
   - `POST /auth/refresh` rotates the refresh token and the old one stops working. If Swift and JS both hold the refresh token, whichever refreshes second presents a revoked token and the user is logged out. Coordinating the two runtimes (shared keychain lock, single writer) is fragile and fails silently on a locked device.
3. Therefore, the app mints a dedicated **intent token** from JS (where the session lives) and hands it to Swift. Swift only presents it; it never refreshes anything.

## Design
- **Endpoint:** `POST /auth/intent-token`, authenticated with the normal Bearer JWT. No `X-Guild-ID`, no body. The user must also still have a session (`user_sessions` row): an access JWT stays valid for up to 15 minutes after logout, and without this check it could mint a token after the user logged out. Returns `201 {"intent_token": "gbi_...", "expires_in": 2592000}`.
- **Scope:** the token is accepted only on `POST /groceries`. The auth middleware rejects it with `403` on every other method/path before looking it up. That includes `/auth/intent-token`, so an intent token cannot mint another one, and `/auth/logout`. A leaked token can add grocery entries to guilds the user is in. It cannot read, delete, or change anything.
- **Guild checks:** the request still sends `X-Guild-ID`. The same user-in-guild and bot-in-guild checks as the JWT path run (one shared helper), so revoking the user's guild membership or kicking the bot still blocks the token.
- **Lifetime:** fixed 30 days from minting, not extended on use. The app re-mints from JS while it has a live session; expiry bounds the damage if the device is lost without a logout.
- **One token per user:** enforced by a unique index on `intent_tokens.discord_user_id`. Minting is a single upsert (`INSERT ... ON CONFLICT (discord_user_id) DO UPDATE`), so the new token replaces the old one atomically and concurrent mints for the same user always leave exactly one row, the last write.
- **Demo accounts excluded:** `POST /auth/intent-token` returns `403` for the demo user (`DEMO_USER_DISCORD_ID`), whether it authenticated through `/auth/demo-login` or a real Discord login. The demo account is shared across devices and demo sessions are meant to be short (no refresh token), so a 30-day token for it doesn't fit. The check runs at mint time only; the app should hide the Siri feature for demo sessions.
- **Revocation:** `POST /auth/token` (new login) and `POST /auth/logout` delete the user's intent tokens. `POST /auth/refresh` does not touch them, since refresh happens constantly and must not break Siri. There is no `DELETE` endpoint; logout or re-mint covers it.
- **Storage:** only the SHA-256 hex hash of the full plaintext token is stored (`intent_tokens.token_hash`, unique index), the same scheme as refresh tokens. A database leak does not yield usable tokens.
- **Rate limiting:** the token lookup runs inside the existing rate limiter keyed by client IP, so floods of invalid `gbi_` tokens are limited before they reach the database. Once the token is found, the request identifier becomes the Discord user ID and the request also counts against the per-user limit it shares with the user's JWT requests.

## Why the `gbi_` prefix?
1. **Routing:** both token types arrive as `Authorization: Bearer ...`. JWTs always start with `eyJ` (base64 of `{"`), so a `gbi_` prefix tells the middleware which verifier to use without trying JWT parsing first and without a new header or auth scheme.
2. **Log redaction:** a fixed prefix makes the token easy to match and mask in logs and crash reports.
3. **Secret scanning:** a distinctive prefix plus fixed-length base64url body (`gbi_[A-Za-z0-9_-]{43}`) can be registered as a custom secret-scanning pattern, so a token committed or pasted somewhere is detectable.

## Known limitations
- **One device per user:** because of the unique index, minting on a second device (for example an iPad) replaces the first device's token, and Siri on the first device stops working until it mints again. This matches the existing one-session-per-user behaviour of `POST /auth/token`. If we want to support multiple devices per user, this may be replaced in the future: for example, a per-device key (such as a `device_id` column) with the unique index on `(discord_user_id, device_id)` and the upsert keyed on both, plus a decision on whether login should still revoke every device's token.
