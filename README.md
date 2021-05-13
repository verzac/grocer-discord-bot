# GroceryBot

[Add this bot to your server](https://discord.com/oauth2/authorize?client_id=815120759680532510&permissions=2048&scope=bot) | [View us at top.gg](https://top.gg/bot/815120759680532510)

GroceryBot allows you to maintain a grocery list for your server/guild/household. You can add and remove entries as you see fit.

~~Because switching to Google Shopping List just to check your grocery list while you're communicating on Discord is a waste of time~~

# Commands

**!grohelp**: Get help!

**!gro \<name\>**: Adds an item to your grocery list.

**!groremove \<n\>**: Removes item #n from your grocery list.

**!groremove \<n\> \<m\> \<o\>...**: Removes item #n, #m, and #o from your grocery list. You can chain as many items as you want.

**!grolist**: List all the groceries in your grocery list

**!groclear**: Clears your grocery list

**!groedit \<n\> \<new name\>**: Updates item #n to a new name/entry

**!grodeets \<n\>**: Views the full detail of item #n (e.g. who made the entry)

**!grohere:**: Attaches a self-updating grocery list to the current channel.

**!groreset**: When you want to clear all of your data from this bot.

```
!grobulk
eggs
Soap 1pc
Liquid soap 500 ml
```

These 3 new items will be added to your existing grocery list!

# Issues & Problems with GroceryBot?

Log an issue here and someone will get back to you: [GitHub Issue](https://github.com/verzac/grocer-discord-bot/issues/new)

NEW: We finally have a support server for GroceryBot, woohoo! You'll be able to ask questions and provide feedback for the bot - straight to the development team! [Click here to join](https://discord.gg/rBjUaZyskg)

# Privacy Policy

"No entry, no data, no problem" and "We don't run analytics or any kind of dodgy stuff because that's expensive"

## Data

When you use a GroceryBot command, we store the following data required for the bot to function properly:

- Your server's ID (because each server has their own grocery list)
- Your server's channel IDs (which isn't human-readable) - this is used so that GroBot knows where to send updates (currently used by !grohere).
- The ID of the user who inputted each grocery entry into your grocery list
- The grocery entry itself (duh)
- When grocery entries are updated (deleted entries are deleted permanently and immediately)

We also keep logs of when an error occurs. This log is automatically disposed of within 14 days.

## Who has access to your data?

Only the maintainer(s) of this project (i.e. [@verzac](https://github.com/verzac) unless otherwise specified) have access to your grocery list because the bot currently lives in his server. And frankly, we don't really want to know what you're going to buy tomorrow.

## Removing your data

While we're sad to see you go, removing your data is as easy as running `!groreset` and removing the bot afterwards. This utility function ensures that all your data is removed from our database.

The only minor exception to the immediate "no entry, no data, no problem" rule is the error logs: error logs are automatically deleted within 14 days (as opposed to immediately). If you've never had an error occur, you shouldn't have this problem.

## Changes to this policy

_Wow, you're actually reading this part. Congratulations, you care about your data!_

While we don't have an official changelog yet where you can view changes to the policy other than GitHub, rest assured we'll try to communicate any changes ASAP to you (maybe by doing the whole "You can't use GroceryBot until you accept this new privacy policy," dance, but that's still to be decided).

# Development

It runs on SQLite and has auto DB migration setup, so you should be set to go immediately.

`go run main.go`

You may need the following env vars set up:

- GROCER_BOT_TOKEN - your Discord bot token
- GROCER_BOT_DSN - DSN target, by default it's set to "db/gorm.db"
