package config

import "os"

func GetN8NApiJWT() string {
	return os.Getenv("N8N_API_JWT")
}

func GetN8NWebhookIngredients() string {
	return os.Getenv("N8N_WEBHOOK_INGREDIENTS")
}

func IsN8NEnabled() bool {
	return GetN8NApiJWT() != "" && GetN8NWebhookIngredients() != ""
}
