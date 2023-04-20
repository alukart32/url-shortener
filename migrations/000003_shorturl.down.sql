ALTER TABLE "shorturls"
    DROP CONSTRAINT "original_uniq";

DROP INDEX IF EXISTS "original_idx";