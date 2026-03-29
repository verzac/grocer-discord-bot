package defaults

import "github.com/bwmarrin/discordgo"

var (
	DefaultListLabelOption = &discordgo.ApplicationCommandOption{
		Type:         discordgo.ApplicationCommandOptionString,
		Name:         "list-label",
		Description:  "Label for your custom grocery list.",
		Required:     false,
		Autocomplete: true,
	}
	DefaultAllListOption = &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "all",
		Description: "Display all of your grocery lists instead of the one denoted in list-label.",
		Required:    false,
	}
)

// ListLabelFromSlashOptions returns the string value of the default list-label option, or "" if absent.
func ListLabelFromSlashOptions(options []*discordgo.ApplicationCommandInteractionDataOption) string {
	for _, option := range options {
		if option.Name == DefaultListLabelOption.Name {
			return option.StringValue()
		}
	}
	return ""
}
