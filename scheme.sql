CREATE TABLE "product"(
    "id" BIGINT NOT NULL,
    "name" VARCHAR(255) NOT NULL
);

ALTER TABLE "product" ADD PRIMARY KEY ("id");
ALTER TABLE "product" ADD CONSTRAINT "product_name_unique" UNIQUE ("name");

CREATE TABLE "version"(
    "id" BIGINT NOT NULL,
    "product_id" BIGINT NOT NULL,
    "version" VARCHAR(255) NOT NULL,
    "file_path" VARCHAR(255) NOT NULL,
    "checksum" VARCHAR(255) NOT NULL
);

ALTER TABLE "version" ADD PRIMARY KEY("id");
ALTER TABLE "version" ADD CONSTRAINT  "version_product_id_foreign" FOREIGN KEY ("product_id") REFERENCES "product" ("id"); 
