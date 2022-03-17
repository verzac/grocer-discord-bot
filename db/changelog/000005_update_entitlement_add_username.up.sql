-- well this was a mistake! make sure there's no data in the tables
DROP INDEX `idx_guild_registrations_guild_id`;
DROP TABLE `guild_registrations`;
DROP TABLE `registration_entitlements`;

-- add the tables, with entitlements not using user_id as PK this time
CREATE TABLE `guild_registrations` (
  `id` integer,
  `guild_id` text not null,
  `registration_entitlement_id` integer not null REFERENCES 'registration_entitlements'('id'),
  `created_at` datetime not null,
  `expires_at` datetime,
  PRIMARY KEY (`id`)
);
CREATE INDEX `idx_guild_registrations_guild_id` ON `guild_registrations`(`guild_id`);

CREATE TABLE `registration_entitlements` (
  `id` integer,
  `user_id` text unique,
  `username` text,
  `username_discriminator` text,
  `expires_at` datetime,
  `external_id` text,
  `external_id_type` text,
  `max_redemption` integer not null,
  'registration_tier_id' integer not null REFERENCES 'registration_tiers'('id'),
  PRIMARY KEY (`id`)
);
CREATE INDEX `idx_registration_entitlements_user_id` ON `registration_entitlements`(`user_id`);
