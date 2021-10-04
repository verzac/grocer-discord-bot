-- grocery_lists definition

CREATE TABLE `grocery_lists` (
  `id` integer,`guild_id` text NOT NULL,
  `list_label` text NOT NULL,
  `fancy_name` text,
  `created_at` datetime,
  `updated_at` datetime,
  PRIMARY KEY (`id`)
);

CREATE INDEX `idx_grocery_lists_list_label` ON `grocery_lists`(`list_label`);
CREATE INDEX `idx_grocery_lists_guild_id` ON `grocery_lists`(`guild_id`);