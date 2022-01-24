package handlers

import "github.com/bwmarrin/discordgo"

func (m *MessageHandlerContext) OnHelp() error {
	version := m.grobotVersion
	subCmd := m.commandContext.ArgStr
	var grohelpMsgEmbed *discordgo.MessageEmbed
	switch subCmd {
	case "beta":
		grohelpMsgEmbed = betaGrohelpMessageEmbed
	default:
		grohelpMsgEmbed = stableGroHelpMessageEmbed
	}
	grohelpMsgEmbed.Title = "GroceryBot " + version
	return m.replyWithEmbed(grohelpMsgEmbed)
}

var (
	stableGroHelpMessageEmbed = &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "<command>:<grocery list label>",
				Value: "[NEW - Multiple grocery lists] Runs <command> on a specific grocery list.\nExample: `!gro:amazon PS5` - adds PS5 to your server's grocery list with the label \"amazon\".",
			},
			{
				Name:  "!grolist new <list label> <fancy display name - optional>",
				Value: "[NEW - Multiple grocery lists] Creates a new grocery list for your server.\nExample: `!grolist new amazon My Amazon Shopping List` - creates a new grocery list; usable through `!gro:amazon your stuff`.",
			},
			{
				Name:  "!grolist:<label> delete",
				Value: "[NEW - Multiple grocery lists] Delete your custom grocery list.\nExample: `!grolist:amazon delete` deletes the grocery list with the label \"amazon\" from your server.\n!grolist also comes with other utility functions - just type `!grolist help`.",
			},
			{
				Name:  "!grohelp",
				Value: "Get help!",
			},
			{
				Name:  "!gro <name>",
				Value: "Adds an item to your grocery list.\nExample: `!gro Chicken katsu` - adds chicken katsu to your grocery list.",
			},
			{
				Name:  "!groremove <n> <m> <o>...",
				Value: "Removes item number #n, #m, and #o from your grocery list. You can chain as many items as you want.\nExample: `!groremove 1 2` - removes item #1 and #2.",
			},
			{
				Name:  "!groremove <item name>",
				Value: "Removes an item which contains <item name> from your grocery list. The item name is case-insensitive. This will delete the first item on your list that contains <new item>.\nExample: `!groremove katsu` - removes \"Chicken katsu\"",
			},
			{
				Name:  "!grolist",
				Value: "List all the groceries in your grocery list.",
			},
			{
				Name:   "!grohere",
				Value:  "Attaches a self-updating grocery list for your grocery list to the current channel.",
				Inline: true,
			},
			{
				Name:   "!grohere all",
				Value:  "Pretty much `!grohere`, except that it displays ALL of your grocery lists.",
				Inline: true,
			},
			{
				Name:  "!groclear",
				Value: "Clears your grocery list.",
			},
			{
				Name:  "!groedit <n> <new name>",
				Value: "Updates item #n to a new name/entry.\nExample: `!groedit 1 Katsudon` - edits item #1 to have the entry Katsudon.",
			},
			{
				Name:  "!groreset",
				Value: "When you want to clear all of your data from this bot. See our privacy policy at https://grocerybot.net/privacy-policy",
			},
			{
				Name: "!grobulk",
				Value: `
Adds multiple items which are separated by newlines.
Example:
` + "```" + `
!grobulk
Chicken 500g
Soap 50ml
Salt
` + "```",
			},
		},
		Description: `
**Release Note:**
:wave: Slash commands are here!

As of April 2022, Discord is effectively deprecating the usage of message commands (the commands you type as a message in your chat box - e.g.` + "`!grohelp`" + `) with their newer "slash commands".

The TL:DR; of this change is that all GroceryBot commands (except !grobulk, because Discord doesn't have the technology to support multi-line input) will have its slash command counterpart. For example, ` + "`!grohelp`" + ` will eventually be replaced by ` + "`/grohelp`" + `.

We understand that this change might not be what you want, so GroceryBot will keep supporting message commands until Discord cuts off our access to your message commands (or, alternatively, you can host an unverified version of GroceryBot to keep using message commands - Discord is only removing message commands for verified bots).

On a bright side, our slash commands now have Autocomplete so that you'll get suggestions for the grocery entries / lists that you want to edit / delete while typing in your commands!

Thank you, and please let us know in our Discord server if you have feedback on our new slash commands! :smile:

[Get Support](https://discord.com/invite/rBjUaZyskg) | [Vote for us at top.gg](https://top.gg/bot/815120759680532510) | [Web](https://grocerybot.net)
	`,
	}

	betaGrohelpMessageEmbed = stableGroHelpMessageEmbed
)
