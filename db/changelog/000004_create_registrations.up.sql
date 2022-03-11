CREATE TABLE `registration_tiers` (
  `id` integer,
  `name` text not null,
  `max_grocery_entry` integer,
  `max_grocery_list` integer,
  PRIMARY KEY (`id`)
);

CREATE TABLE `guild_registrations` (
  `id` integer,
  `guild_id` text not null,
  `registration_entitlement_user_id` text not null REFERENCES 'registration_entitlements'('id'),
  `created_at` datetime not null,
  `expires_at` datetime,
  PRIMARY KEY (`id`)
);

CREATE INDEX `idx_guild_registrations_guild_id` ON `guild_registrations`(`guild_id`);

CREATE TABLE `registration_entitlements` (
  `user_id` text,
  `expires_at` datetime,
  `external_id` text,
  `external_id_type` text,
  `max_redemption` integer not null,
  'registration_tier_id' integer not null REFERENCES 'registration_tiers'('id'),
  PRIMARY KEY (`user_id`)
);

INSERT INTO `registration_tiers` (`id`, `name`, `max_grocery_entry`, `max_grocery_list`)
VALUES 
  (1, 'GroPatron', 150, 5),
  (2, 'GroPremium', 300, 10);