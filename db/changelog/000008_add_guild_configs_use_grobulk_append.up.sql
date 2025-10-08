ALTER TABLE `guild_configs` ADD COLUMN
  `use_grobulk_append` boolean DEFAULT false;

-- set all existing guild configs to false (we assume that they're using the old grobulk behaviour, so make this opt-in for existing users)
UPDATE `guild_configs` SET `use_grobulk_append` = true;
