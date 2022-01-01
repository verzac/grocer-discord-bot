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
	_, err := m.sess.ChannelMessageSendEmbed(m.commandContext.ChannelID, grohelpMsgEmbed)
	return err
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
:shopping_bags: We're proud to announce the release of multiple grocery lists for GroceryBot 2.1! 

With this new feature, you will be able to maintain multiple grocery lists within your server - perfect for when you want to have separate shopping lists for your Sunday market trip and Amazon shopping spree.

All commands support this new feature with the following syntax:` + "`<command>:<list-label>` (e.g. `!gro:amazon PS5`)" + `. We can't wait for you to try out this new feature, so please let us know in our Discord server if you see any issue (or potential improvement that we can do)!

**What's next?**
Discord is planning to deprecate (i.e. effectively get rid of) the usage of traditional "message commands" on April 2022, so we're planning to add support for slash commands very soon. Don't worry though: the change should be minimal, other than replacing "!" with "/" (e.g. "/grolist").

[Get Support](https://discord.com/invite/rBjUaZyskg) | [Vote for us at top.gg](https://top.gg/bot/815120759680532510) | [Web](https://grocerybot.net)
	`,
	}

	betaGrohelpMessageEmbed = stableGroHelpMessageEmbed
)
