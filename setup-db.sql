#CREATE DATABASE spoto OWNER spoto  ENCODING 'UTF-8' template template0;

CREATE TABLE "public"."media" (
    "id" serial NOT NULL,
    "iid" text NOT NULL,
    "document" jsonb NOT NULL,
    "created_at" timestamp NOT NULL,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."subscriptions" (
    "id" serial NOT NULL,
    "name" text NOT NULL,
    "lat" real NOT NULL,
    "long" real NOT NULL,
    "radius" int NOT NULL,
    PRIMARY KEY ("id")
);

ALTER TABLE "public"."media"
  ADD COLUMN "subscription_id" int NOT NULL,
  ADD FOREIGN KEY ("subscription_id") REFERENCES "public"."subscriptions"("id");