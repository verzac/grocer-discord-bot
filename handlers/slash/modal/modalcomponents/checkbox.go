package modalcomponents

import "github.com/bwmarrin/discordgo"

var _ discordgo.MessageComponent = &Checkbox{}

type Checkbox struct{}

// MarshalJSON implements [discordgo.MessageComponent].
func (c *Checkbox) MarshalJSON() ([]byte, error) {
	panic("unimplemented")
}

// Type implements [discordgo.MessageComponent].
func (c *Checkbox) Type() discordgo.ComponentType {
	panic("unimplemented")
}
