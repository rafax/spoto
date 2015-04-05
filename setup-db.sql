CREATE TABLE "public"."notifications" (
    "id" serial NOT NULL,
    "iid" text NOT NULL,
    "object" text NOT NULL,
    "changed_aspect" text NOT NULL,
    "changed_time" timestamp NOT NULL,
    PRIMARY KEY ("id")
);

CREATE TABLE "public"."subscriptions" (
    "id" serial NOT NULL,
    "name" text NOT NULL,
    "subscriptionId" text NOT NULL,
    PRIMARY KEY ("id")
);

ALTER TABLE "public"."notifications"
  ADD COLUMN "subscription_id" int NOT NULL,
  ADD FOREIGN KEY ("subscription_id") REFERENCES "public"."subscriptions"("id");