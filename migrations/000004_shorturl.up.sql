ALTER TABLE "shorturls"
    ADD COLUMN "deleted" BOOLEAN;

ALTER TABLE "shorturls" ALTER
    COLUMN "deleted" SET NOT NULL;

ALTER TABLE "shorturls" ALTER
    COLUMN "deleted" SET DEFAULT false;