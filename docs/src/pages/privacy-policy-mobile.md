---
sidebar_position: 3
---

# Privacy Policy — GroceryBot App

**Last updated:** May 2, 2026

**Operator ("we"):** Benjamin Tanone, the maintainer of GroceryBot and ([github.com/verzac](https://github.com/verzac)).

**Contact:** [hello@benjamintanone.com](mailto:hello@benjamintanone.com)

This policy explains how **GroceryBot App** — the mobile client for the GroceryBot service at [grocerybot.net](https://grocerybot.net) — handles your information. The app connects to **api.grocerybot.net** so you can manage grocery lists tied to Discord servers ("guilds") you belong to.

**Read alongside:** the [GroceryBot website privacy policy](https://grocerybot.net/privacy-policy/), which covers the bot itself, server-side storage, human access to data, backend performance aggregates, error-log retention (~14 days), and `/groreset`. This document adds the details specific to GroceryBot App: in-app OAuth login, the local offline cache, and planned mobile analytics and crash tooling. **If anything here conflicts with the website privacy policy, the website policy takes precedence.**

## Where data lives

Primary GroceryBot data — including the databases that hold your grocery lists — is processed and stored on infrastructure in **Singapore**. See the [GroceryBot website policy](https://grocerybot.net/privacy-policy/) for maintainer access and operational practices.

Mobile analytics and crash-reporting vendors may process certain events or reports **outside Singapore**, depending on which services are integrated. Retention follows each vendor's dashboard settings and contracts.

---

## Information the app uses

- **Account and sign-in.** You sign in with **Discord** via OAuth2. Discord processes your account under [Discord's Privacy Policy](https://discord.com/privacy). The app requests only the scopes it needs for identity and server membership (currently `identify`, `guilds`, and `guilds.members.read`).

- **Grocery and server data.** Content and metadata synced with GroceryBot's servers — server IDs, channel IDs where applicable, list names, grocery entries, authoring user IDs, timestamps, and similar. See the website policy for the authoritative list.

- **Data on your device.** Session tokens are kept in **secure storage**. A local copy of your lists and guild metadata is cached to support **offline viewing**. Your last selected server may also be stored locally.

- **Analytics (usage and performance).** We may use services such as **Google Analytics** or similar tools for **anonymous, aggregate** product insight — screens viewed, timings, app version, coarse device/OS. Grocery list contents are **not** intentionally captured or tied to ad profiles.

- **Crash and error reports.** Crash reports sent to our chosen provider intentionally include your **Discord user ID** — the primary account identifier the app uses — plus technical context (stack traces, device/OS level, timestamps) so incidents can be matched to affected users. Grocery list contents and other in-app user content are **not** deliberately included in crash payloads.

- **API request logs.** Calls to **api.grocerybot.net** are logged with standard request information (such as IP address, request path, status code, and timestamp) for security, debugging, and abuse prevention. These logs are not used to build advertising profiles.

Human access to backend data follows the stance set out on [grocerybot.net](https://grocerybot.net/privacy-policy/): no routine browsing of the production database, except **with your written consent** for support or debugging.

## How we use this information

To operate GroceryBot, to improve reliability and usability through aggregated analytics and crash diagnosis, and to address abuse where needed. We do **not** sell your grocery content to unrelated parties for advertising.

## Sharing with third parties

We do **not** sell personal information for marketing, and we do **not** pass grocery entries to unrelated third parties for resale or profiling.

Recipients that may process data on our behalf:

- **Discord** — OAuth sign-in; Discord's terms apply during that step.
- **Analytics providers** (e.g. Google Analytics) — see the [Google Privacy Policy](https://policies.google.com/privacy).
- **Crash-reporting providers** — Discord user ID plus technical diagnostics, as described above.
- **App stores, Expo / EAS (updates), and hosting providers (AWS)** — ordinary operational telemetry provided by those platforms.

---

## Removing your data

- **Individual entries.** Delete in the app; while online, the deletion syncs to the server and the entry is removed permanently, consistent with backend practice.

- **`/groreset` (Discord).** Clears essentially everything for that server in GroceryBot's database, **except** individual users' Patreon tier records, which are retained so patrons can continue to register and link servers.

- **Removing GroceryBot from a server.** Documented on grocerybot.net as part of leaving the service entirely.

- **Signing out.** Clears local session tokens. Your data on GroceryBot's servers is **not** deleted — sign in again with the same Discord account and your lists will be there. Cached groceries may remain on your device until you clear the app's cache or uninstall it.

See also "Removing your data" on [grocerybot.net](https://grocerybot.net/privacy-policy/), including backend error-log disposal (~14 days).

Backend analytics described on grocerybot.net is fully anonymous aggregates. Mobile analytics inside this app is **planned** under the restrictions above and not yet shipped.

---

## Children

The service is not directed at minors. It supports Discord communities and tooling aimed at teenagers and adults, and Discord's own minimum age applies independently. Parents and guardians are responsible for supervising minors' Discord use.

## Security

Protect your device and your Discord account — anyone with access to either controls the linked data.

---

## Regional notices

Privacy laws vary by country and region. We have not added jurisdiction-specific annexes (such as GDPR-style rights wording for the EEA/UK, California CPRA disclosures, or Brazilian LGPD notices), because GroceryBot App is a hobby-scale service without commercial ad-tech or storefront-driven compliance requirements.

This policy already covers Singapore storage, operator identity, contact email, the minors stance, the `/groreset` Patreon carve-out, analytics and crash scopes, and links to the [grocerybot.net policy](https://grocerybot.net/privacy-policy/). If you are in a region that gives you statutory rights — such as access, erasure, portability, objection, restriction, or the right to complain to a supervisory authority — you can exercise them by emailing **[hello@benjamintanone.com](mailto:hello@benjamintanone.com)**.

A business postal address is intentionally not published. Formal regulatory notices can be initiated by email.

---

## Contact for privacy requests

For any privacy question or rights request — access, correction, deletion, or anything else covered above — email **[hello@benjamintanone.com](mailto:hello@benjamintanone.com)**. Please mention your Discord user ID (or the username/handle you signed in with) so we can locate the right account.

---

## Changes

We update **Last updated** above when the text changes materially.
