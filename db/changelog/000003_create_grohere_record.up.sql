CREATE TABLE `grohere_records` (
  `id` integer,
  `guild_id` text not null,
  `grocery_list_id` integer REFERENCES 'grocery_lists'('id'),
  `grohere_channel_id` text not null,
  `created_at` datetime,
  `updated_at` datetime, 
  `grohere_message_id` text not null,
  PRIMARY KEY (`id`)
);

CREATE INDEX `idx_grohere_records_guild_id` ON `grohere_records`(`guild_id`);
CREATE INDEX `idx_grohere_records_grocery_list_id` ON `grohere_records`(`grocery_list_id`);
