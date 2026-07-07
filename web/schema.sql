-- D1 schema for the waitlist + funnel events.
-- Apply with:
--   wrangler d1 execute radius-tools --file=schema.sql            (local dev)
--   wrangler d1 execute radius-tools --file=schema.sql --remote   (production)

CREATE TABLE IF NOT EXISTS waitlist (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT NOT NULL UNIQUE,
  failure_signature TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  detail TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_events_name_created ON events (name, created_at);
