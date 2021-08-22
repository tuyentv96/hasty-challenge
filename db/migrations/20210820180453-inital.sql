
-- +migrate Up
CREATE TABLE IF NOT EXISTS "jobs" (
"id" serial PRIMARY KEY,
"object_id" integer NOT NULL,
"status" text NOT NULL,
"start_time" timestamp(6),
"end_time" timestamp(6),
"message" TEXT,
"created_at" timestamp(6) NOT NULL DEFAULT timezone('utc'::text, now())
);

-- +migrate Down
DROP TABLE IF EXISTS "jobs";

