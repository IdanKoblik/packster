CREATE TABLE "user"(
    "id" SERIAL NOT NULL,
    "display_name" VARCHAR(255) NOT NULL,
    "last_login" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    "created_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);
ALTER TABLE
    "user" ADD PRIMARY KEY("id");
CREATE TABLE "host"(
    "id" SERIAL NOT NULL,
    "url" VARCHAR(255) NOT NULL,
    "host_type" VARCHAR(255) NOT NULL
);
ALTER TABLE
    "host" ADD PRIMARY KEY("id");
ALTER TABLE
    "host" ADD CONSTRAINT "host_url_unique" UNIQUE("url");
CREATE TABLE "project"(
    "id" SERIAL NOT NULL,
    "host" INTEGER NOT NULL,
    "repository" INTEGER NOT NULL,
    "owner" INTEGER NOT NULL,
    "created_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);
ALTER TABLE
    "project" ADD PRIMARY KEY("id");
CREATE INDEX "project_repository_index" ON
    "project"("repository");
CREATE TABLE "product"(
    "id" SERIAL NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "project" INTEGER NOT NULL,
    "created_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);
ALTER TABLE
    "product" ADD PRIMARY KEY("id");
ALTER TABLE
    "product" ADD CONSTRAINT "product_name_project_unique" UNIQUE("name", "project");
CREATE TABLE "token"(
    "id" SERIAL NOT NULL,
    "value" VARCHAR(255) NOT NULL,
    "expiry" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL,
    "owner" INTEGER NOT NULL,
    "created_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL
);
ALTER TABLE
    "token" ADD PRIMARY KEY("id");
CREATE TABLE "token_access"(
    "token" INTEGER NOT NULL,
    "product" INTEGER NOT NULL
);
ALTER TABLE
    "token_access" ADD PRIMARY KEY("token", "product");
CREATE TABLE "auth"(
    "id" SERIAL NOT NULL,
    "account" INTEGER NOT NULL,
    "username" VARCHAR(255) NOT NULL,
    "sso_id" INTEGER NOT NULL,
    "host" INTEGER NOT NULL
);
ALTER TABLE
    "auth" ADD PRIMARY KEY("id");
CREATE TABLE "permission"(
    "account" INTEGER NOT NULL,
    "project" INTEGER NOT NULL,
    "can_download" BOOLEAN NOT NULL,
    "can_upload" BOOLEAN NOT NULL,
    "can_delete" BOOLEAN NOT NULL
);
ALTER TABLE
    "permission" ADD PRIMARY KEY("account", "project");
CREATE TABLE "org"(
    "id" INTEGER NOT NULL,
    "host" INTEGER NOT NULL
);
ALTER TABLE
    "org" ADD PRIMARY KEY("id", "host");
CREATE TABLE "version"(
    "id" SERIAL NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "path" VARCHAR(255) NOT NULL,
    "checksum" VARCHAR(255) NOT NULL,
    "product" INTEGER NOT NULL
);
ALTER TABLE
    "version" ADD PRIMARY KEY("id");
ALTER TABLE
    "version" ADD CONSTRAINT "version_name_product_unique" UNIQUE("name", "product");
ALTER TABLE
    "org" ADD CONSTRAINT "org_host_foreign" FOREIGN KEY("host") REFERENCES "host"("id");
ALTER TABLE
    "token" ADD CONSTRAINT "token_owner_foreign" FOREIGN KEY("owner") REFERENCES "user"("id");
ALTER TABLE
    "auth" ADD CONSTRAINT "auth_account_foreign" FOREIGN KEY("account") REFERENCES "user"("id");
ALTER TABLE
    "token_access" ADD CONSTRAINT "token_access_token_foreign" FOREIGN KEY("token") REFERENCES "token"("id");
ALTER TABLE
    "project" ADD CONSTRAINT "project_host_foreign" FOREIGN KEY("host") REFERENCES "host"("id");
ALTER TABLE
    "permission" ADD CONSTRAINT "permission_project_foreign" FOREIGN KEY("project") REFERENCES "project"("id");
ALTER TABLE
    "auth" ADD CONSTRAINT "auth_host_foreign" FOREIGN KEY("host") REFERENCES "host"("id");
ALTER TABLE
    "project" ADD CONSTRAINT "project_owner_foreign" FOREIGN KEY("owner") REFERENCES "user"("id");
ALTER TABLE
    "token_access" ADD CONSTRAINT "token_access_product_foreign" FOREIGN KEY("product") REFERENCES "product"("id");
ALTER TABLE
    "version" ADD CONSTRAINT "version_product_foreign" FOREIGN KEY("product") REFERENCES "product"("id");
ALTER TABLE
    "product" ADD CONSTRAINT "product_project_foreign" FOREIGN KEY("project") REFERENCES "project"("id");
ALTER TABLE
    "permission" ADD CONSTRAINT "permission_account_foreign" FOREIGN KEY("account") REFERENCES "user"("id");