CREATE TABLE IF NOT EXISTS `waitlist_ios` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `discord_user_id` text NOT NULL,
  `email` text NOT NULL,
  `name` text,
  `created_at` datetime
);

CREATE INDEX `idx_waitlist_ios_discord_user_id` ON `waitlist_ios`(`discord_user_id`);
