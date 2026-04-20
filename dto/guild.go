package dto

type UserGuild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type UserGuildsResponse struct {
	Guilds []UserGuild `json:"guilds"`
}
