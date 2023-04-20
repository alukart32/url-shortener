ALTER TABLE "shorturls"
    ADD CONSTRAINT "original_uniq" UNIQUE ("original");

CREATE INDEX "original_idx" ON "shorturls" ("original");