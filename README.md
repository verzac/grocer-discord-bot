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

## Data

When you use a GroceryBot command, we store the following data required for the bot to function properly:

- Your server's ID (because each server has their own grocery list)
- The ID of the user who inputted each grocery entry into your grocery list
- The grocery entry itself (duh)
- When grocery entries are updated (deleted entries are deleted permanently and immediately)

We also keep logs of when an error occurs. This log is automatically disposed of within 14 days.

## Who has access to your data?

Only the maintainer(s) of this project (i.e. [@verzac](https://github.com/verzac) unless otherwise specified) have access to your grocery list because the bot currently lives in his server. And frankly, we don't really want to know what you're going to buy tomorrow.

# Development

It runs on SQLite and has auto DB migration setup, so you should be set to go immediately.

`go run main.go`

You may need the following env vars set up:

- GROCER_BOT_TOKEN - your Discord bot token
- GROCER_BOT_DSN - DSN target, by default it's set to "db/gorm.db"
