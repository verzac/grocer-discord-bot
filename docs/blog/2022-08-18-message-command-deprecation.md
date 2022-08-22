---
slug: message-command-deprecation
title: Commands Through Plain Messages Will Stop Working
authors: verzac
tags: [grocerybot, general]
---

**From 31 August 2022, commands that are sent through plain messages that do not mention GroceryBot will stop working.** This is a policy change by Discord. You can either mention GroceryBot, or use the new slash commands that we've developed.\*

(Press **Read More** for more info)

<!--truncate-->

\*except `!grobulk` - this is on our roadmap as we are waiting for Discord to finalise multi-line forms.

In any case...

Slash commands are finally here!

TL:DR; !gro -> /gro, !grolist -> /grolist, and every other commands except !grobulk because Discord doesn't have the tech to process multi-line slash commands yet.

As of 31 August 2022, Discord will be deprecating message commands (i.e. doing things with bots by sending messages - e.g. !gro blah), and are replacing it with "Slash Commands" (e.g. /gro blah). Message commands will not work from September 2022 onwards, unless if you mention GroceryBot directly.

While I disagree with Discord's nuclear option (and the fact that slash commands themselves aren't 100% finished), GroceryBot's slash commands comes with a lot of cool, nifty features such as auto-completion.

As always, I know this transition is a sudden one to everyone, but we do have to comply with Discord's policy. If you've got any suggestions, please feel free to hop onto our Discord server and mention me (@verzac) - we want to make sure that our users can get the best experience possible!

## Links & References

[Discord's Announcement](https://support-dev.discord.com/hc/en-us/articles/4404772028055-Message-Content-Privileged-Intent-FAQ)

## FAQ

### I miss the old commands! Slash commands suck!

Don't worry, we're probably on the same boat. If you miss the old experience, you can always mention `@GroceryBot`. e.g.

```
@GroceryBot !gro:my-list chicken
```

### My slash commands don't work!

Re-invite the bot through the link on this site (at the bottom, in the footer).

Due to a botched invite link ([Github issue here](https://github.com/verzac/grocer-discord-bot/issues/14#issuecomment-1095888210)), some of our users who used that botched invite link were unable to use slash commands. This should have been fixed - if it doesn't work please hop on to our support server!

**Your data is safe: your data will persist even if you kick the bot out of your server** (_until_ you run `/groreset`). This is an intentional design decision to make re-invites easy.
