package announcement

import (
	"context"
	"fmt"

	"github.com/verzac/grocer-discord-bot/models"
	"go.uber.org/zap"
)

const (
	CurrentAnnouncementVersion = 3
	AnnouncementMessage        = "📱 **GroceryBot's app is now available on iOS too!** Check your grocery lists across all your Discord servers in one tap — no more navigating between servers and channels. Works offline too!\n\n[For Android](<https://play.google.com/store/apps/details?id=net.grocerybot.app>) | [For iOS](<https://apps.apple.com/us/app/grocerybot-app/id6780291729?itscg=30200&itsct=apps_box_link&mttnsubad=6780291729>) | [Blog Post](<https://grocerybot.net/blog/new-grocerybot-app>)"
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
