-- Create index "idx_user_identity_sso_provider_id_user_id" to table: "user_identities"
CREATE UNIQUE INDEX "idx_user_identity_sso_provider_id_user_id" ON "user_identities" ("sso_provider_id", "user_id");
