---
sidebar_position: 3
---

# Privacy Policy

"No entry, no data, no problem," and "We don't run analytics or any kind of dodgy stuff because that's expensive."

## Who are the maintainers?

Heyo 👋! I'm [@verzac](https://github.com/verzac), and I'm currently the only official maintainer of this bot. The bot runs on my AWS server in Singapore, and no-one but the "maintainers" have access to the underlying infrastructure (including the aforementioned server) that powers the bot.

## Who has access to your data?

GroceryBot processes your data given by you to the bot through its commands so that it can provide you with its intended services. Our policy is that NO HUMANS are allowed to manually look at GroceryBot's database, unless if you have specifically provided written consent (e.g. to debug an issue that you are having with the bot).

It is also important to note that we do aggregate and collect anonymised performance data. For example, we record how long GroceryBot takes to serve up a request/command within a particular period on average. Rest assured though, these data are fully anonymous (i.e. from those data alone, no-one can figure out who people are and what they're using the bot for).

## What kind of data do we collect?

When you use a GroceryBot command, we store the following data required for the bot to function properly:

- Your server's ID (because each server has their own grocery list)
- Your server's channel IDs (which isn't human-readable) - this is used so that GroBot knows where to send updates (currently used by !grohere).
- The ID of the user who inputted each grocery entry into your grocery list
- The grocery entry itself (duh)
- When grocery entries are updated (deleted entries are deleted permanently and immediately)
- Your created grocery lists' names and labels.

We also keep logs of when an error occurs. This log is automatically disposed of within 14 days.

## Removing your data

While we're sad to see you go, removing your data is as easy as running `!groreset` or `/groreset` and removing the bot afterwards. This utility function ensures that all your data is removed from our database.

The only minor exception to the immediate "no entry, no data, no problem" rule is the error logs: error logs are automatically deleted within 14 days (as opposed to immediately). If you've never had an error occur, you shouldn't have this problem.

## Changes to this policy

_Wow, you're actually reading this part. Congratulations, you care about your data!_

This site will pretty much always be updated with the latest privacy policy. Practically, this policy will never receive any major revision, unless if there are new features being introduced. If you'd like to be updated when new versions of the privacy policy is out, please hop on to our [support server on Discord](https://discord.com/invite/rBjUaZyskg) and ping `@verzac`; he'll add you to an announcement list.
