package announcement

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (s *AnnouncementServiceImpl) AugmentMessageWithAnnouncement(ctx context.Context, guildID string, message string) (string, error) {
	guildConfig, err := s.guildconfigRepo.Get(guildID)
	if err != nil {
		s.logger.Error("Failed to get guild config", zap.Error(err), zap.String("guildID", guildID))
		return message, nil // Return original message on error (non-critical failure)
	}

	// If guild config is nil, guild not initialized yet, return original message
	if guildConfig == nil {
		return message, nil
	}

	// Check if we need to show announcement
	if guildConfig.LastAnnouncementVersion < CurrentAnnouncementVersion {
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

		s.logger.Info("Announcement shown to guild",
			zap.String("guildID", guildID),
			zap.Int("announcementVersion", CurrentAnnouncementVersion))

		return augmentedMessage, nil
	}

	return message, nil
}
