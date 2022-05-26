package slash

import "github.com/bwmarrin/discordgo"

var modalCommandHandlers = map[string]func(i *discordgo.InteractionCreate) *discordgo.InteractionResponseData{
	"grorecipe": func(i *discordgo.InteractionCreate) *discordgo.InteractionResponseData {
		return &discordgo.InteractionResponseData{
			CustomID: "grorecipe_" + i.Interaction.Member.User.ID,
			Title:    "Make a new recipe",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "prettyName",
							Label:       "What do you want to call your recipe?",
							Placeholder: "Enter a fancy name for your recipe! (Max 255 characters)",
							Required:    false,
							MaxLength:   255,
							Style:       discordgo.TextInputShort,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "prettyName",
							Label:       "What do you want to put in your recipe?",
							Placeholder: "Separate each entry with a new-line, just like !grobulk",
							Required:    false,
							MaxLength:   255,
							Style:       discordgo.TextInputParagraph,
						},
					},
				},
			},
		}
	},
}

type ModalHandler struct {
}
