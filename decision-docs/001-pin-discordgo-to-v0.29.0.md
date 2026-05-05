# Pin Discordgo to v0.29.0 (forked version under verzac/discordgo-new-components)

Author: verzac (github.com/verzac)

## Why is this important?
1. Repo updated discordgo to point to `verzac/discordgo-new-components@master`, which is `bwmarrin/discordgo@master` + 1 extra commit to add new components
  - It was forked because the discordgo repo has no active maintainers, and I wanted to use the new Checkbox components from Discord.
  - Plus, I wanted to take advantage of new feature improvements on Discordgo@master.
2. However, `master` contains changes which prevents the bot from reconnecting to Discord properly (see Appendix). This behaviour does not exist in v0.29.0 - the bot was running on v0.29.0 for 2 years without needing restarts.
3. Therefore, I decided to move the new components to v0.29.0 - the latest stable version of discordgo - instead of master.

## Appendix: Opus 4.6 Investigation
I found the root cause. It's a **deadlock** introduced by the Op9 (Invalid Session) handler change in discordgo master.

## Root Cause

The key change is in `wsapi.go`. In v0.29.0, the Op9 handler simply re-identified on the existing connection:

```go
// v0.29.0 - Op9 handler
err = s.identify()  // only uses s.wsMutex, NOT s.Lock()
```

In master (your fork), it was changed to close and reconnect:

```go
// master - Op9 handler
s.CloseWithCode(websocket.CloseServiceRestart)  // calls s.Lock()!
s.reconnect()
```

The problem is that `onEvent` is called from **two contexts**:

1. **From `listen()` goroutine** (normal operation) -- lock is NOT held, works fine
2. **From within `Open()`** during the handshake -- lock IS held, **deadlocks**

Here's the deadlock path, which happens on every reconnection attempt where the session has expired:

```
listen() detects disconnect
  → Close()                       // preserves sessionID, sequence, resumeGatewayURL
  → reconnect()
    → Open()
      → s.Lock()                  // ← acquires session lock
      → sends Op6 Resume          // (because sessionID/sequence are preserved)
      → reads response
      → s.onEvent(...)            // Discord responds with Op9 (session expired)
        → CloseWithCode()
          → s.Lock()              // ← DEADLOCK! same goroutine, non-reentrant mutex
```

The bot hangs forever at this point -- it can't reconnect, can't process events, nothing. Go's `sync.RWMutex` is not reentrant, so the second `Lock()` call from the same goroutine blocks indefinitely.

This is a very common scenario: whenever Discord restarts a gateway or the session expires between disconnect and reconnect, Discord responds to the Resume (Op6) with Op9 Invalid Session, triggering the deadlock.

## Fix

The simplest fix: **rebase your fork on `v0.29.0` instead of `master`**. Since your fork only adds one struct, cherry-pick that single commit onto v0.29.0. This avoids the deadlock and also avoids several other breaking API changes in master (e.g., `ChannelMessagesPinned` signature change, `MessageComponentInteractionDataResolved` rename, `TextInput.Required` type change from `bool` to `*bool`).

If you'd rather stay on master and fix it in your fork, the fix is to handle Op9 during the `Open()` handshake before it reaches `onEvent`. Want me to help with either approach?