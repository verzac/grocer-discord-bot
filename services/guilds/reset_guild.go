package guilds

import (
	"context"

	"github.com/verzac/grocer-discord-bot/models"
	"gorm.io/gorm"
)

func (s *GuildsServiceImpl) ResetGuild(ctx context.Context, guildID string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if r := tx.Delete(&models.GroceryEntry{}, "guild_id = ?", guildID); r.Error != nil {
			return r.Error
		}
		if r := tx.Delete(&models.GuildConfig{}, "guild_id = ?", guildID); r.Error != nil {
			return r.Error
		}
		if r := tx.Delete(&models.GroceryList{}, "guild_id = ?", guildID); r.Error != nil {
			return r.Error
		}
		if r := tx.Delete(&models.GrohereRecord{}, "guild_id = ?", guildID); r.Error != nil {
			return r.Error
		}
		if r := tx.Delete(&models.GuildRegistration{}, "guild_id = ?", guildID); r.Error != nil {
			return r.Error
		}
		if r := tx.Delete(&models.ApiClient{}, "client_id = ?", guildID); r.Error != nil {
			return r.Error
		}
		return nil
	})
}
