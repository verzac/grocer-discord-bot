package auth

import "strings"

const guildScopePrefix = "guild:"

// GetScopeForGuild helper func to enforce consistent formatting
func GetScopeForGuild(guildID string) string {
	return guildScopePrefix + guildID
}

func getAllowedGuildIDsFromScope(scope string) (out []string) {
	out = []string{}
	scopes := strings.Split(scope, ",")
	for _, s := range scopes {
		if strings.HasPrefix(s, guildScopePrefix) {
			tokens := strings.SplitN(s, ":", 2)
			if len(tokens) == 2 {
				out = append(out, tokens[1])
			}
		}
	}
	return
}

func GetGuildIDFromScope(scope string) string {
	guildIDs := getAllowedGuildIDsFromScope(scope)
	if len(guildIDs) == 0 {
		return ""
	}
	return guildIDs[0]
}
