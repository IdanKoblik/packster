CREATE TABLE `products`(
    `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(255) NOT NULL,
    `group_name` VARCHAR(255) NOT NULL,
    `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP());
ALTER TABLE
    `products` ADD UNIQUE `products_name_group_name_unique`(`name`, `group_name`);
CREATE TABLE `product_versions`(
    `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `product_id` BIGINT NOT NULL,
    `name` VARCHAR(255) NOT NULL,
    `path` TEXT NOT NULL,
    `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP());
CREATE TABLE `auth`(
    `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `type` ENUM('token', 'gitlab_user') NOT NULL,
    `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP());
CREATE TABLE `gitlab_users`(
    `id` BIGINT NOT NULL,
    `gitlab_user_id` BIGINT NOT NULL,
    `username` VARCHAR(255) NULL,
    PRIMARY KEY(`id`)
);
ALTER TABLE
    `gitlab_users` ADD UNIQUE `gitlab_users_gitlab_user_id_unique`(`gitlab_user_id`);
CREATE TABLE `product_permissions`(
    `principal_id` BIGINT NOT NULL,
    `product_id` BIGINT NOT NULL,
    `can_download` BOOLEAN NULL,
    `can_upload` BOOLEAN NULL,
    `can_remove` BOOLEAN NULL,
    `is_maintainer` BOOLEAN NULL,
    PRIMARY KEY(`principal_id`)
);
ALTER TABLE
    `product_permissions` ADD PRIMARY KEY(`product_id`);
ALTER TABLE
    `gitlab_users` ADD CONSTRAINT `gitlab_users_id_foreign` FOREIGN KEY(`id`) REFERENCES `auth`(`id`);
ALTER TABLE
    `product_permissions` ADD CONSTRAINT `product_permissions_product_id_foreign` FOREIGN KEY(`product_id`) REFERENCES `products`(`id`);
ALTER TABLE
    `product_versions` ADD CONSTRAINT `product_versions_product_id_foreign` FOREIGN KEY(`product_id`) REFERENCES `products`(`id`);
ALTER TABLE
