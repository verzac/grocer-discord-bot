package slash

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func groceryListLabelHash(listLabel string) string {
	sum := sha1.Sum([]byte(listLabel))
	return hex.EncodeToString(sum[:])
}

func parseListHashFromCheckboxCustomID(customID string) string {
	const prefix = "list_hash="
	idx := strings.Index(customID, prefix)
	if idx < 0 {
		return ""
	}
	rest := customID[idx+len(prefix):]
	end := strings.Index(rest, ":group=")
	if end < 0 {
		return ""
	}
	return rest[:end]
}

type groremoveRawModalSubmit struct {
	InteractionID string
	Token         string
	GuildID       string
	ChannelID     string
	Member        *discordgo.Member
	User          *discordgo.User
	ListHash      string
	Values        []string
}

func walkModalSubmitComponents(raw json.RawMessage, visit func(customID string, values []string)) {
	var node struct {
		Type        int               `json:"type"`
		Component   json.RawMessage   `json:"component"`
		Components  []json.RawMessage `json:"components"`
		CustomID    string            `json:"custom_id"`
		Values      []string          `json:"values"`
	}
	if err := json.Unmarshal(raw, &node); err != nil {
		return
	}
	switch node.Type {
	case 1:
		for _, c := range node.Components {
			walkModalSubmitComponents(c, visit)
		}
	case 18:
		if len(node.Component) > 0 {
			walkModalSubmitComponents(node.Component, visit)
		}
	case 22:
		if node.CustomID != "" {
			visit(node.CustomID, node.Values)
		}
	}
}

func parseGroremoveModalSubmitFromInteractionRaw(raw []byte) (*groremoveRawModalSubmit, bool, error) {
	var env struct {
		ID        string `json:"id"`
		Token     string `json:"token"`
		Type      int    `json:"type"`
		GuildID   string `json:"guild_id"`
		ChannelID string `json:"channel_id"`
		Member    *discordgo.Member `json:"member"`
		User      *discordgo.User   `json:"user"`
		Data      struct {
			CustomID   string            `json:"custom_id"`
			Components []json.RawMessage `json:"components"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, false, err
	}
	if env.Type != int(discordgo.InteractionModalSubmit) || env.Data.CustomID != "groremove" {
		return nil, false, nil
	}
	out := &groremoveRawModalSubmit{
		InteractionID: env.ID,
		Token:         env.Token,
		GuildID:       env.GuildID,
		ChannelID:     env.ChannelID,
		Member:        env.Member,
		User:          env.User,
	}
	var listHash string
	visit := func(customID string, values []string) {
		if h := parseListHashFromCheckboxCustomID(customID); h != "" && listHash == "" {
			listHash = h
		}
		out.Values = append(out.Values, values...)
	}
	for _, c := range env.Data.Components {
		walkModalSubmitComponents(c, visit)
	}
	out.ListHash = listHash
	return out, true, nil
}

func dedupeStringsPreserveOrder(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
