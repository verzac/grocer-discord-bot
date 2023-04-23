package handlers

import (
	"strings"

	"github.com/pkg/errors"
)

// OnConfig handles !groconfig set use_ephemeral true
func (m *MessageHandlerContext) OnConfig() error {
	argStrTokens := strings.SplitN(m.commandContext.ArgStr, " ", 4)
	if len(argStrTokens) != 3 {
		return m.onError(errors.New("Invalid command. Must follow the format of /config set <thingamabobs>"))
	}
	return nil
}
