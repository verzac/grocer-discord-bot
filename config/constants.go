package config

import (
	"os"
	"strings"
)

const (
	GrobotVersionLocal = "local"
)

func GetDefaultMaxGroceryEntriesPerServer() int {
	return 100
}

func GetDefaultMaxGroceryListsPerServer() int {
	return 3
}

func IsMaintenanceMode() bool {
	return os.Getenv("GROCER_BOT_MAINTENANCE_MODE") == "true"
}

func GetListOfGuildIDsToIgnoreForMetrics() []string {
	return strings.Split(os.Getenv("GROCER_BOT_IGNORE_GUILD_ID_LIST"), ",")
}
