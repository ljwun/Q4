-- Create "sso_providers" table
CREATE TABLE "sso_providers" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_sso_providers_name" UNIQUE ("name")
);
-- Create index "idx_sso_providers_deleted_at" to table: "sso_providers"
CREATE INDEX "idx_sso_providers_deleted_at" ON "sso_providers" ("deleted_at");
-- Drop index "idx_users_username" from table: "users"
DROP INDEX "idx_users_username";
-- Create "user_identities" table
CREATE TABLE "user_identities" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "sso_provider_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "identity" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_user_identities_sso_provider" FOREIGN KEY ("sso_provider_id") REFERENCES "sso_providers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_user_identities_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_user_identities_deleted_at" to table: "user_identities"
CREATE INDEX "idx_user_identities_deleted_at" ON "user_identities" ("deleted_at");
-- Create index "idx_user_identity_sso_provider_id_identity" to table: "user_identities"
CREATE UNIQUE INDEX "idx_user_identity_sso_provider_id_identity" ON "user_identities" ("sso_provider_id", "identity");
