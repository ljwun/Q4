-- Drop index "idx_user_identity_sso_provider_id_identity" from table: "user_identities"
DROP INDEX "idx_user_identity_sso_provider_id_identity";
-- Drop index "idx_user_identity_sso_provider_id_user_id" from table: "user_identities"
DROP INDEX "idx_user_identity_sso_provider_id_user_id";
-- Create index "idx_user_identity_sso_provider_id_identity" to table: "user_identities"
CREATE UNIQUE INDEX "idx_user_identity_sso_provider_id_identity" ON "user_identities" ("sso_provider_id", "identity") WHERE (deleted_at IS NULL);
-- Create index "idx_user_identity_sso_provider_id_user_id" to table: "user_identities"
CREATE UNIQUE INDEX "idx_user_identity_sso_provider_id_user_id" ON "user_identities" ("sso_provider_id", "user_id") WHERE (deleted_at IS NULL);
