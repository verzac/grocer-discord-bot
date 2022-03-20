package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func getSpecialThanksMsg(mentions []string) string {
	const msgFormat = "Special thanks to %s for your patronage!"
	if len(mentions) == 0 {
		return fmt.Sprintf(msgFormat, "you guys")
	} else {
		mentionsString := ""
		for i, m := range mentions {
			if i == 0 {
				mentionsString += m
			} else if i == len(mentions)-1 {
				mentionsString += ", and" + m
			} else {
				mentionsString += ", " + m
			}
		}
		return fmt.Sprintf(msgFormat, mentionsString)
	}
}

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
	registrationCtx := m.GetRegistrationContext()
	if !registrationCtx.IsDefault {
		grohelpMsgEmbed.Description = fmt.Sprintf(`
			**YOUR BENEFITS**
			Max grocery entries: %d
			Max grocery lists: %d
			%s
			`, registrationCtx.MaxGroceryEntriesPerServer, registrationCtx.MaxGroceryListsPerServer, getSpecialThanksMsg(registrationCtx.RegistrationsOwnersMention),
		) + grohelpMsgEmbed.Description
	}
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
**WHAT'S NEW**
:tada: Become a GroPatron through my Patreon page (link below) to get access to higher limits and support the bot's development!

:mega: On April 30 2022, Discord will be removing GroceryBot's ability to see messages that do not directly mention it. Therefore: 
1. **Make sure you mention @GroceryBot before running your commands** (do this: ` + "`" + `@GroceryBot !gro chicken` + "`" + `, not this: ` + "`" + `!gro chicken` + "`" + `); OR
2. Use our new slash commands (which comes with nifty auto-completion)!

[Get Support](https://discord.com/invite/rBjUaZyskg) | [Patreon](https://www.patreon.com/verzac) | [Vote for us at top.gg](https://top.gg/bot/815120759680532510) | [Web](https://grocerybot.net)
	`,
	}

	betaGrohelpMessageEmbed = stableGroHelpMessageEmbed
)
