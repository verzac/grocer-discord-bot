package dto

type DiscordError struct {
	Code   int `json:"code"`
	Errors struct {
		Content struct {
			Errors []struct {
				Code    string
				Message string
			} `json:"_errors"`
		} `json:"content"`
	} `json:"errors"`
}
