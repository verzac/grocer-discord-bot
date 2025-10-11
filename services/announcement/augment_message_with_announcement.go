package announcement

import (
	"context"
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
)

const (
	CurrentAnnouncementVersion = 1
	AnnouncementMessage        = "Psst... You can now use the `/grobulk` command to **edit your existing grocery lists in one go** - no more running `/groremove` and `/groclear` once you're done. Try it out now using `/config set use_grobulk_replace=True`!"
)

func (s *AnnouncementServiceImpl) AugmentMessageWithAnnouncement(ctx context.Context, guildID string, message string) (string, error) {
	guildConfig, err := s.guildconfigRepo.Get(guildID)
	if err != nil {
		s.logger.Error("Failed to get guild config", zap.Error(err), zap.String("guildID", guildID))
		return message, nil // Return original message on error (non-critical failure)
	}

	if guildConfig == nil {
		guildConfig = &models.GuildConfig{
			GuildID:                 guildID,
			LastAnnouncementVersion: 0,
		}
	}

	lastAnnouncementVersion := guildConfig.LastAnnouncementVersion

	// Check if we need to show announcement
	if lastAnnouncementVersion < CurrentAnnouncementVersion {
		// Append announcement to message
		augmentedMessage := fmt.Sprintf("%s\n\n%s", message, AnnouncementMessage)

		// Update guild config to mark announcement as shown
		guildConfig.LastAnnouncementVersion = CurrentAnnouncementVersion
		if err := s.guildconfigRepo.Put(guildConfig); err != nil {
			s.logger.Error("Failed to update guild config with announcement version",
				zap.Error(err),
				zap.String("guildID", guildID),
				zap.Int("announcementVersion", CurrentAnnouncementVersion))
			// Still return augmented message even if save fails
		}

		return augmentedMessage, nil
	}

	return message, nil
}
