package slash

import (
	"encoding/json"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/require"
)

func TestParseListHashFromCheckboxCustomID(t *testing.T) {
	t.Parallel()
	cid := "groremove:list_hash=deadbeef:group=2"
	require.Equal(t, "deadbeef", parseListHashFromCheckboxCustomID(cid))
	require.Equal(t, "", parseListHashFromCheckboxCustomID("groremove:group=0"))
}

func TestGroceryListLabelHash_DefaultList(t *testing.T) {
	t.Parallel()
	h := groceryListLabelHash("")
	require.Equal(t, "da39a3ee5e6b4b0d3255bfef95601890afd80709", h)
}

func TestParseGroremoveModalSubmitFromInteractionRaw(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{
  "id": "interaction-id",
  "token": "interaction-token",
  "type": 5,
  "guild_id": "guild-1",
  "channel_id": "channel-1",
  "member": {
    "user": {"id": "user-1", "username": "u", "discriminator": "0"}
  },
  "data": {
    "custom_id": "groremove",
    "components": [
      {
        "type": 18,
        "label": "Items 1–2",
        "component": {
          "type": 22,
          "custom_id": "groremove:list_hash=abc123:group=0",
          "values": ["1", "3"]
        }
      },
      {
        "type": 1,
        "components": [
          {
            "type": 22,
            "custom_id": "groremove:list_hash=abc123:group=1",
            "values": ["2"]
          }
        ]
      }
    ]
  }
}`)

	parsed, ok, err := parseGroremoveModalSubmitFromInteractionRaw(raw)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "interaction-id", parsed.InteractionID)
	require.Equal(t, "interaction-token", parsed.Token)
	require.Equal(t, "guild-1", parsed.GuildID)
	require.Equal(t, "channel-1", parsed.ChannelID)
	require.Equal(t, "abc123", parsed.ListHash)
	require.Equal(t, []string{"1", "3", "2"}, parsed.Values)
}

func TestParseGroremoveModalSubmitFromInteractionRaw_NotModalSubmit(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{"type": 2, "data": {"custom_id": "groremove"}}`)
	_, ok, err := parseGroremoveModalSubmitFromInteractionRaw(raw)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestParseGroremoveModalSubmitFromInteractionRaw_WrongCustomID(t *testing.T) {
	t.Parallel()
	raw := json.RawMessage(`{"type": 5, "data": {"custom_id": "grobulk"}}`)
	_, ok, err := parseGroremoveModalSubmitFromInteractionRaw(raw)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestWalkModalSubmitComponents_CheckboxOnly(t *testing.T) {
	t.Parallel()
	var seen []string
	walkModalSubmitComponents(json.RawMessage(`{
		"type": 22,
		"custom_id": "groremove:list_hash=ff:group=0",
		"values": ["9"]
	}`), func(customID string, values []string) {
		require.Equal(t, "groremove:list_hash=ff:group=0", customID)
		seen = append(seen, values...)
	})
	require.Equal(t, []string{"9"}, seen)
}

func TestDedupeStringsPreserveOrder(t *testing.T) {
	t.Parallel()
	require.Equal(t, []string{"1", "2"}, dedupeStringsPreserveOrder([]string{"1", "1", "2", "1"}))
}

func TestModalSubmitCustomID(t *testing.T) {
	t.Parallel()
	i := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionModalSubmit,
			Data: discordgo.ModalSubmitInteractionData{
				CustomID: "groremove",
			},
		},
	}
	cid, ok := modalSubmitCustomID(i)
	require.True(t, ok)
	require.Equal(t, "groremove", cid)
}
