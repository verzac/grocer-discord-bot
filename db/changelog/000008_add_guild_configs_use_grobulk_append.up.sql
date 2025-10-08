ALTER TABLE `guild_configs` ADD COLUMN
  `use_grobulk_append` boolean DEFAULT false;

-- set all existing guild configs to true - we assume that they're using the old grobulk behaviour, so make this opt-in through /config for existing users
UPDATE `guild_configs` SET `use_grobulk_append` = true;
