-- EXPORTED FROM DBEAVER

-- grocery_entries definition

CREATE TABLE IF NOT EXISTS "grocery_entries" (
  `id` integer,
  `created_at` datetime,
  `updated_at` datetime,
  `item_desc` text NOT NULL,
  `guild_id` text NOT NULL, 
  `creator_id` text, 
  `updated_by_id` text,
  PRIMARY KEY (`id`)
);

CREATE INDEX IF NOT EXISTS `idx_grocery_entries_guild_id` ON `grocery_entries`(`guild_id`);

-- guild_configs definition

CREATE TABLE IF NOT EXISTS `guild_configs` (
  `guild_id` text,
  `grohere_channel_id` text,
  `created_at` datetime,
  `updated_at` datetime, 
  `grohere_message_id` text,
  PRIMARY KEY (`guild_id`)
);