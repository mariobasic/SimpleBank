DROP TABLE IF EXISTS "verify_emails" CASCADE;

ALTER TABLE "users" DROP COLUMN IF EXISTS "is_email_verified";

ALTER TABLE "sessions" RENAME COLUMN "expired_at" to "expires_at";