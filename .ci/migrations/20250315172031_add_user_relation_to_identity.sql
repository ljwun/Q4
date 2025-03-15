-- Modify "user_identities" table
ALTER TABLE "user_identities" DROP CONSTRAINT "fk_user_identities_user", ADD
 CONSTRAINT "fk_users_identities" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
