CREATE TABLE `user_sessions` (
  `id` integer,
  `created_at` datetime,
  `updated_at` datetime,
  `deleted_at` datetime,
  `discord_user_id` text NOT NULL,
  `discord_access_token` text NOT NULL,
  `discord_refresh_token` text NOT NULL,
  `discord_token_expiry` datetime NOT NULL,
  `refresh_token_hash` text NOT NULL,
  `refresh_token_expiry` datetime NOT NULL,
  PRIMARY KEY (`id`)
);

CREATE UNIQUE INDEX `idx_user_sessions_discord_user_id` ON `user_sessions`(`discord_user_id`);
CREATE INDEX `idx_user_sessions_refresh_token_hash` ON `user_sessions`(`refresh_token_hash`);
CREATE INDEX `idx_user_sessions_deleted_at` ON `user_sessions`(`deleted_at`);
