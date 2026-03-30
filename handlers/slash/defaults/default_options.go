package defaults

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

const DefaultEntryOptionName = "entry"

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

// EntryFromSlashOptions returns the trimmed string value of the entry option, or "" if absent.
func EntryFromSlashOptions(options []*discordgo.ApplicationCommandInteractionDataOption) string {
	for _, option := range options {
		if option.Name == DefaultEntryOptionName {
			return strings.TrimSpace(option.StringValue())
		}
	}
	return ""
}
