CREATE TABLE IF NOT EXISTS `user_sessions` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `discord_user_id` text NOT NULL,
  `refresh_token_hash` text NOT NULL,
  `refresh_token_expiry` datetime NOT NULL,
  `created_at` datetime,
  `updated_at` datetime
);

CREATE INDEX `idx_user_sessions_discord_user_id` ON `user_sessions`(`discord_user_id`);
CREATE INDEX `idx_user_sessions_refresh_token_hash` ON `user_sessions`(`refresh_token_hash`);
