ALTER TABLE `user_sessions` ADD COLUMN `encrypted_discord_access_token` text NOT NULL DEFAULT '';
ALTER TABLE `user_sessions` ADD COLUMN `encrypted_discord_refresh_token` text NOT NULL DEFAULT '';
ALTER TABLE `user_sessions` ADD COLUMN `discord_token_expiry` datetime;
