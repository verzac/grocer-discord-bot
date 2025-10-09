package guildconfig

import (
	"github.com/bwmarrin/discordgo"
	"github.com/verzac/grocer-discord-bot/models"
	"github.com/verzac/grocer-discord-bot/repositories"
	"go.uber.org/zap"
)

func (s *GuildConfigServiceImpl) InitialiseGuildConfig(sess *discordgo.Session) {
	tx := s.db.Begin()
	defer tx.Rollback()

	guildConfigRepository := &repositories.GuildConfigRepositoryImpl{DB: tx}
	globalFlagRepository := &repositories.GlobalFlagRepositoryImpl{DB: tx}

	// Check if we've already initialized guild configs
	hasInitFlag, err := globalFlagRepository.GetFlag("has_init_guild_configs")
	if err != nil {
		s.logger.Error("Failed to check has_init_guild_configs flag", zap.Error(err))
		return
	}

	if hasInitFlag == "true" {
		s.logger.Info("Guild configs already initialized, skipping.")
		return
	}

	s.logger.Info("Initialising guild config for existing servers (if needed).")

	// Get all guilds the bot is in
	guilds := sess.State.Guilds

	initializedCount := 0
	for _, guild := range guilds {
		// Check if guild config already exists
		existingConfig, err := guildConfigRepository.Get(guild.ID)
		if err != nil {
			s.logger.Error("Failed to check existing guild config", zap.String("guildID", guild.ID), zap.Error(err))
			break
		}

		// If no config exists, create a default one
		if existingConfig == nil {
			existingConfig = &models.GuildConfig{
				GuildID: guild.ID,
			}

			initializedCount++
		}

		// update the grobulk append flag to preserve legacy behaviours
		existingConfig.UseGrobulkAppend = true

		// persist the final one
		if err := guildConfigRepository.Put(existingConfig); err != nil {
			s.logger.Error("Failed to create default guild config", zap.String("guildID", guild.ID), zap.Error(err))
			break
		}
	}

	s.logger.Info("Guild config initialization completed", zap.Int("initializedCount", initializedCount), zap.Int("totalGuilds", len(guilds)))

	// Set the flag to indicate we've completed initialization
	if err := globalFlagRepository.SetFlag("has_init_guild_configs", "true"); err != nil {
		s.logger.Error("Failed to set has_init_guild_configs flag", zap.Error(err))
		return
	}

	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit guild config transaction", zap.Error(err))
		return
	}

	s.logger.Info("Guild config initialization completed successfully.")
}
