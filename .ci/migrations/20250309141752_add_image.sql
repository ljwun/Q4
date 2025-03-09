-- Create "images" table
CREATE TABLE "images" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "uploader_id" uuid NOT NULL,
  "url" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_images_uploader" FOREIGN KEY ("uploader_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_images_deleted_at" to table: "images"
CREATE INDEX "idx_images_deleted_at" ON "images" ("deleted_at");
