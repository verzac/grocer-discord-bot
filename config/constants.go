package config

import "os"

func GetDefaultMaxGroceryEntriesPerServer() int {
	return 100
}

func GetDefaultMaxGroceryListsPerServer() int {
	return 3
}

func IsMaintenanceMode() bool {
	return os.Getenv("GROCER_BOT_MAINTENANCE_MODE") == "true"
}
