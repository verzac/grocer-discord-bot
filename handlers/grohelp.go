package handlers

import "github.com/bwmarrin/discordgo"

func (m *MessageHandlerContext) OnHelp(version string) error {
	_, err := m.sess.ChannelMessageSendEmbed(m.msg.ChannelID, &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "!grohelp",
				Value: "Get help!",
			},
			{
				Name:  "!gro <name>",
				Value: "Adds an item to your grocery list.",
			},
			{
				Name:  "!groremove <n> <m> <o>...",
				Value: "Removes item #n, #m, and #o from your grocery list. You can chain as many items as you want.",
			},
			{
				Name:  "!grolist",
				Value: "List all the groceries in your grocery list.",
			},
			{
				Name:  "!groclear",
				Value: "Clears your grocery list.",
			},
			{
				Name:  "!groedit <n> <new name>",
				Value: "Updates item #n to a new name/entry.",
			},
			{
				Name:  "!grodeets <n>",
				Value: "Views the full detail of item #n (e.g. who made the entry).",
			},
			{
				Name:  "!grohere",
				Value: "Attaches a self-updating grocery list to the current channel.",
			},
			{
				Name:  "!groreset",
				Value: "When you want to clear all of your data from this bot.",
			},
			{
				Name: "!grobulk",
				Value: `
Adds multiple items which are separated by newlines. For example:
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
!grohere is here! Now, you can attach and pin a message to a specific channel that GroBot will automatically update as you update your grocery list. No more typing multiple !grolist commands just to see your updated grocery list! :tada:

[Get Support](https://discord.com/invite/rBjUaZyskg) | [Vote for us at top.gg](https://top.gg/bot/815120759680532510) | [GitHub](https://github.com/verzac/grocer-discord-bot)
		`,
	})
	return err
}
