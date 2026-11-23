CREATE TABLE IF NOT EXISTS `intent_tokens` (
  `id` integer PRIMARY KEY AUTOINCREMENT,
  `discord_user_id` text NOT NULL,
  `token_hash` text NOT NULL,
  `expires_at` datetime NOT NULL,
  `created_at` datetime
);

CREATE INDEX `idx_intent_tokens_discord_user_id` ON `intent_tokens`(`discord_user_id`);
CREATE UNIQUE INDEX `idx_intent_tokens_token_hash` ON `intent_tokens`(`token_hash`);
