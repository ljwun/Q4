-- Create "auction_items" table
CREATE TABLE "auction_items" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "user_id" uuid NULL,
  "title" character varying(255) NOT NULL,
  "description" text NOT NULL,
  "starting_price" integer NOT NULL,
  "current_bid_id" uuid NULL,
  "start_time" timestamptz NOT NULL,
  "end_time" timestamptz NOT NULL,
  "carousels" text[] NULL DEFAULT '{}',
  PRIMARY KEY ("id")
);
-- Create index "idx_auction_items_deleted_at" to table: "auction_items"
CREATE INDEX "idx_auction_items_deleted_at" ON "auction_items" ("deleted_at");
-- Create "bids" table
CREATE TABLE "bids" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "amount" integer NOT NULL,
  "user_id" uuid NOT NULL,
  "auction_item_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_bids_deleted_at" to table: "bids"
CREATE INDEX "idx_bids_deleted_at" ON "bids" ("deleted_at");
-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v7(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "username" character varying(255) NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_users_deleted_at" to table: "users"
CREATE INDEX "idx_users_deleted_at" ON "users" ("deleted_at");
-- Create index "idx_users_username" to table: "users"
CREATE UNIQUE INDEX "idx_users_username" ON "users" ("username");
-- Modify "auction_items" table
ALTER TABLE "auction_items" ADD
 CONSTRAINT "fk_auction_items_current_bid" FOREIGN KEY ("current_bid_id") REFERENCES "bids" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
 CONSTRAINT "fk_auction_items_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "bids" table
ALTER TABLE "bids" ADD
 CONSTRAINT "fk_auction_items_bid_records" FOREIGN KEY ("auction_item_id") REFERENCES "auction_items" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
 CONSTRAINT "fk_bids_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
