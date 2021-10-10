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
:eyes: Why hello there! We haven't had a release recently, but guess what: GroceryBot 2 is here! It comes with a lot of overhaul within the underyling infrastructure which would (hopefully) allow the development team to deliver a bigger and better GroceryBot for you.

Also, this update comes with a not-yet-officially-released feature that a few of us have been waiting for. If you've been lurking in our Discord server's announcement channel, you probably know what this feature is :wink:. You can already start using it if you wish to get ahead of the pack, but be warned that this is "provided as is" (unlike the other commands) and the underlying commands might break/change without notice. We expect the official release date to be a month from now, so get keen on this while we're testing the new feature in production!

[Get Support](https://discord.com/invite/rBjUaZyskg) | [Vote for us at top.gg](https://top.gg/bot/815120759680532510) | [Web](https://grocerybot.net)
		`,
	})
	return err
}
