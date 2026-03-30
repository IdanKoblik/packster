
CREATE TABLE `products`(
                           `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
                           `name` VARCHAR(255) NOT NULL,
                           `group_name` VARCHAR(255) NOT NULL,
                           `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP(),

                           UNIQUE KEY `products_name_group_unique` (`name`, `group_name`)
);

CREATE TABLE `product_versions`(
                                   `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
                                   `product_id` BIGINT NOT NULL,
                                   `name` VARCHAR(255) NOT NULL,
                                   `path` TEXT NOT NULL,
                                   `checksum` VARCHAR(255) NOT NULL,
                                   `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP(),

                                   CONSTRAINT `fk_product_versions_product`
                                       FOREIGN KEY (`product_id`)
                                           REFERENCES `products`(`id`)
                                           ON DELETE CASCADE
);

CREATE TABLE `principals`(
                             `id` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
                             `type` ENUM('token', 'gitlab_user') NOT NULL,
                             `admin` BOOLEAN,
                             `created_at` TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP()
);

CREATE TABLE `api_tokens`(
                             `id` BIGINT NOT NULL,
                             `token_hash` VARCHAR(255) NOT NULL,

                             PRIMARY KEY(`id`),

                             CONSTRAINT `fk_api_tokens_principal`
                                 FOREIGN KEY (`id`)
                                     REFERENCES `principals`(`id`)
                                     ON DELETE CASCADE
);

CREATE TABLE `gitlab_users`(
                               `id` BIGINT NOT NULL,
                               `gitlab_user_id` BIGINT NOT NULL,
                               `username` VARCHAR(255) NULL,

                               PRIMARY KEY(`id`),

                               UNIQUE KEY `gitlab_users_gitlab_user_id_unique` (`gitlab_user_id`),

                               CONSTRAINT `fk_gitlab_users_principal`
                                   FOREIGN KEY (`id`)
                                       REFERENCES `principals`(`id`)
                                       ON DELETE CASCADE
);

CREATE TABLE `product_permissions`(
                                      `principal_id` BIGINT NOT NULL,
                                      `product_id` BIGINT NOT NULL,
                                      `can_download` BOOLEAN DEFAULT FALSE,
                                      `can_upload` BOOLEAN DEFAULT FALSE,
                                      `can_remove` BOOLEAN DEFAULT FALSE,
                                      `is_maintainer` BOOLEAN DEFAULT FALSE,

                                      PRIMARY KEY(`principal_id`, `product_id`),

                                      CONSTRAINT `fk_permissions_principal`
                                          FOREIGN KEY (`principal_id`)
                                              REFERENCES `principals`(`id`)
                                              ON DELETE CASCADE,

                                      CONSTRAINT `fk_permissions_product`
                                          FOREIGN KEY (`product_id`)
                                              REFERENCES `products`(`id`)
                                              ON DELETE CASCADE
);