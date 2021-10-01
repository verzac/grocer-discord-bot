package handlers

import "github.com/bwmarrin/discordgo"

func (m *MessageHandlerContext) OnHelp() error {
	version := m.grobotVersion
	_, err := m.sess.ChannelMessageSendEmbed(m.msg.ChannelID, &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			// {
			// 	Name:  "[NEW - Multiple grocery lists] <command>:<grocery list label>",
			// 	Value: "Runs <command> on a specific grocery list.\nExample: `!gro:amazon PS5` - adds PS5 to your server's grocery list with the label \"amazon\".",
			// },
			// {
			// 	Name:  "[NEW - Multiple grocery lists] !grolist new <list label> <fancy display name - optional>",
			// 	Value: "Creates a new grocery list for your server.\nExample: `!grolist new amazon My Amazon Shopping List` - creates a new grocery list; usable through `!gro:amazon your stuff`.",
			// },
			// {
			// 	Name:  "[NEW - Multiple grocery lists] !grolist:<label> delete",
			// 	Value: "Delete your custom grocery list.\nExample: `!grolist delete amazon` deletes the grocery list with the label \"amazon\" from your server.\n!grolist also comes with other utility functions - just type `!grolist help`.",
			// },
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
				Name:  "!grohere",
				Value: "Attaches a self-updating grocery list to the current channel.",
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
		Title: "GroceryBot " + version,
		Description: `
**Release Note:**
We've added support to !groremove items based on their name alone, so that you don't have to type !grolist and look for the entry # just to delete an item :tada:

We're also looking into :globe_with_meridians: moving off onto a new website with better documentation for GroBot; and :robot: adding support for smart home integration. Get keen!

[Get Support](https://discord.com/invite/rBjUaZyskg) | [Vote for us at top.gg](https://top.gg/bot/815120759680532510) | [Web](https://grocerybot.net)
		`,
	})
	return err
}
