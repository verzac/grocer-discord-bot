CREATE TABLE `api_clients` (
  `id` integer,
  -- `guild_id` text not null,
  `created_at` datetime not null,
  `updated_at` datetime not null,
  `deleted_at` datetime,
  `created_by_id` text not null,
  `client_id` text not null,
  `client_secret` text not null,
  `scope` text not null,
  PRIMARY KEY (`id`)
);

CREATE INDEX `idx_api_keys_scope` ON `api_clients`(`scope`);
CREATE INDEX `idx_api_keys_client_id` ON `api_clients`(`client_id`);
