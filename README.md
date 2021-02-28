# GroceryBot

[Add this bot to your server](https://discord.com/oauth2/authorize?client_id=815120759680532510&permissions=2048&scope=bot)

GroceryBot allows you to maintain a grocery list for your server/guild/household. You can add and remove entries as you see fit.

~~Because switching to Google Shopping List just to check your grocery list while you're communicating on Discord is a waste of time~~

# Commands

**!grohelp**: get help!

**!gro <name>**: adds an item to your grocery list

**!groremove <n>**: removes item #n from your grocery list

**!grolist**: list all the groceries in your grocery list

**!groclear**: clears your grocery list

**!groedit <n> <new name>**: updates item #n to a new name/entry

You can also do **!grobulk** to add your own grocery list. Format:

```
!grobulk
eggs
Soap 1pc
Liquid soap 500 ml
```

These 3 new items will be added to your existing grocery list!

# Development

It runs on SQLite and has auto DB migration setup, so you should be set to go immediately.

`go run main.go`

You may need the following env vars set up:

- GROCER_BOT_TOKEN - your Discord bot token
- GROCER_BOT_DSN - DSN target, by default it's set to "db/gorm.db"
