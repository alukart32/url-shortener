CREATE TABLE "shorturls" (
  "slug" char(7) PRIMARY KEY,
  "user_id" uuid NOT NULL,
  "original" varchar NOT NULL,
  "short" varchar NOT NULL
);

CREATE INDEX "user_id_idx" ON "shorturls" ("user_id");