# GroceryBot

[Add this bot to your server](https://discord.com/oauth2/authorize?client_id=815120759680532510&permissions=2048&scope=bot)

GroceryBot allows you to maintain a grocery list for your server/guild/household. You can add and remove entries as you see fit.

~~Because switching to Google Shopping List just to check your grocery list while you're communicating on Discord is a waste of time~~

# Commands

**!grohelp**: get help!

**!gro \<name\>**: adds an item to your grocery list

**!groremove \<n\>**: removes item #n from your grocery list

**!groremove \<n\> \<m\> \<o\>...**: removes item #n, #m, and #o from your grocery list. You can chain as many items as you want.

**!grolist**: list all the groceries in your grocery list

**!groclear**: clears your grocery list

**!groedit \<n\> \<new name\>**: updates item #n to a new name/entry

**!grodeets \<n\>**: views the full detail of item #n (e.g. who made the entry)

You can also do **!grobulk** to add your own grocery list. Format:

```
!grobulk
eggs
Soap 1pc
Liquid soap 500 ml
```

These 3 new items will be added to your existing grocery list!

# Privacy Policy

"No entry, no data, no problem" and "We don't run analytics or any kind of dodgy stuff because that's expensive"

## Data

When you use a GroceryBot command, we store the following data required for the bot to function properly:

- Your server's ID (because each server has their own grocery list)
- The ID of the user who inputted each grocery entry into your grocery list
- The grocery entry itself (duh)
- When grocery entries are updated (deleted entries are deleted permanently and immediately)

We also keep logs of when an error occurs. This log is automatically disposed of within 14 days.

## Who has access to your data?

Only the maintainer(s) of this project (i.e. [@verzac](https://github.com/verzac) unless otherwise specified) have access to your grocery list because the bot currently lives in his server. And frankly, we don't really want to know what you're going to buy tomorrow.

## Removing your data

While we're sad to see you go, removing your data is as easy as running `!groclear` and removing the bot afterwards. All of the data above is stored against your grocery entries, so if you don't have any grocery entry, your data is not stored at all.

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
