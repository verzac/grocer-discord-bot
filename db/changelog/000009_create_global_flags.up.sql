CREATE TABLE `global_flags` (
  `key` varchar(255) NOT NULL PRIMARY KEY,
  `value` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP
);
