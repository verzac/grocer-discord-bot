---
sidebar_position: 3
---

# Privacy Policy — GroceryBot App

**Last updated:** July 3, 2026

**Who's behind this:** GroceryBot is primarily built and maintained by Benjamin Tanone ([github.com/verzac](https://github.com/verzac)).

**Contact:** [hello@benjamintanone.com](mailto:hello@benjamintanone.com)

This policy covers **GroceryBot App** — the mobile client that connects to [grocerybot.net](https://grocerybot.net) so you can manage grocery lists tied to your Discord servers.

You should also read the [GroceryBot website privacy policy](https://grocerybot.net/privacy-policy/). That one covers the bot itself, server-side storage, who can access data, backend performance metrics, error-log retention (~14 days), and `/groreset`. This document adds what's specific to the app: OAuth login, the offline cache, and planned analytics and crash tooling. If the two ever conflict, the website policy wins.

---

## Where your data lives

Your grocery list data is processed and stored on infrastructure in **Singapore**. The [website policy](https://grocerybot.net/privacy-policy/) has more detail on how things are managed day-to-day.

Analytics and crash-reporting services may process some events **outside Singapore**, depending on the vendor. Retention follows each vendor's own settings and contracts.

---

## What the app collects and why

**Sign-in.** You log in through Discord's OAuth2 flow — Discord handles that under [its own privacy policy](https://discord.com/privacy). We only request the scopes the app actually needs: `identify`, `guilds`, and `guilds.members.read` (mostly to check which Discord server you belong to).

**Grocery and server data.** Everything that syncs with GroceryBot's servers — server IDs, channel IDs where relevant, list names, grocery entries, who added what, timestamps, and so on. The website policy has the full list.

**Data on your device.** Session tokens are stored in secure storage. A local copy of your lists and guild info is cached so you can browse offline (handy for supermarkets with terrible reception). Your last selected server is also stored locally.

**Analytics.** We may use tools like Google Analytics for **anonymous, aggregate** insight — screens visited, load times, app version, rough device/OS info. We don't capture grocery list contents, and none of this feeds ad profiles.

**Crash reports.** When the app crashes, the report includes your **Discord user ID** (that's how we identify accounts) plus technical context like stack traces, device info, and timestamps. Grocery list contents and other personal content are **not** deliberately included.

**API logs.** Calls to **api.grocerybot.net** are logged with standard request info (IP address, path, status code, timestamp) for security, debugging, and abuse prevention — not for ads.

We don't routinely browse the production database. We only look at your data **with your written consent** — typically for support or debugging. More on this in the [website policy](https://grocerybot.net/privacy-policy/).

---

## How we use this information

To run GroceryBot, improve its reliability through aggregate analytics, diagnose crashes, and deal with abuse. We don't sell your grocery data to anyone for advertising.

---

## Who else gets access

We don't sell personal information for marketing, and we don't hand grocery entries to third parties for resale or profiling.

That said, some services process data on our behalf:

- **Discord** — handles OAuth sign-in under its own terms.
- **Analytics providers** (e.g. Google Analytics) — see the [Google Privacy Policy](https://policies.google.com/privacy).
- **Crash-reporting providers** — receive your Discord user ID and technical diagnostics as described above.
- **App stores, Expo/EAS (updates), and hosting (AWS)** — standard operational telemetry from those platforms.

---

## Deleting your data

**Individual entries.** Delete them in the app. While you're online, the deletion syncs to the server and the entry is gone for good.

**`/groreset` on Discord.** Wipes essentially everything for that server in GroceryBot's database, **except** Patreon tier records — those are kept so patrons can continue linking servers. Please reach-out to the maintainer if you'd like to completely erase your data.

**Removing the bot from a server.** Covered on [grocerybot.net](https://grocerybot.net/privacy-policy/).

**Signing out.** Clears your local session tokens but does **not** delete your data on GroceryBot's servers — sign back in and your lists will still be there. Cached data may stick around on your device until you clear the app's storage or uninstall it.

More detail in the "Removing your data" section of the [website policy](https://grocerybot.net/privacy-policy/), including error-log cleanup (~14 days).

Backend analytics on grocerybot.net uses fully anonymous aggregates. In-app mobile analytics is **planned** under the restrictions above but hasn't shipped yet.

---

## Children

GroceryBot isn't aimed at children. It's built for Discord communities of teenagers and adults, and Discord's own minimum-age requirement applies independently. Parents and guardians are responsible for supervising minors' use.

## Security

Keep your device and Discord account secure — anyone with access to either can reach the data linked to them.

---

## Your rights

Privacy laws vary by region. We haven't added formal jurisdiction-specific annexes (GDPR for the EEA/UK, CPRA for California, LGPD for Brazil, etc.) — GroceryBot is a hobby-scale project without commercial ad-tech or data-driven revenue.

That said, this policy already covers the essentials: where data is stored, who we are, how to reach us, the stance on minors, the `/groreset` Patreon carve-out, what analytics and crash reports include, and links to the [website policy](https://grocerybot.net/privacy-policy/).

**If your region gives you statutory rights** — access, rectification, erasure, portability, restriction of processing, objection, or the right to lodge a complaint with a supervisory authority — you can exercise them by emailing **[hello@benjamintanone.com](mailto:hello@benjamintanone.com)**. We'll do our best to respond promptly.

We don't publish a business postal address. Formal regulatory correspondence can be initiated by email.

---

## Contact

For any privacy question or request — access, correction, deletion, or anything else — email **[hello@benjamintanone.com](mailto:hello@benjamintanone.com)**. Please include your Discord user ID or username so we can find the right account.

---

## Changes

We update the **Last updated** date at the top when the text changes materially.
