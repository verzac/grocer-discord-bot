CREATE TABLE `api_keys` (
  `id` integer,
  -- `guild_id` text not null,
  `created_at` datetime not null,
  `created_by_id` text not null,
  `api_key_hashed` text not null,
  `scope` text not null
);

CREATE INDEX `idx_api_keys_scope` ON `api_keys`(`scope`);
CREATE INDEX `idx_api_keys_api_key_hashed` ON `api_keys`(`api_key_hashed`);
