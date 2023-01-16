CREATE TABLE IF NOT EXISTS relay_scores (
  pubkey TEXT,
  url TEXT,
  score INT,

  UNIQUE (pubkey, url)
);

CREATE TABLE IF NOT EXISTS events (
  id TEXT PRIMARY KEY,
  body BLOB
);

CREATE TABLE IF NOT EXISTS profile_events (
  pubkey TEXT PRIMARY KEY,
  id TEXT,
  date INTEGER,

  UNIQUE (pubkey, id)
);

CREATE INDEX IF NOT EXISTS idx_profile_events_date ON profile_events (date);

CREATE TABLE IF NOT EXISTS reply_events (
  root TEXT,
  id TEXT,
  date INTEGER,

  UNIQUE (root, id)
);

CREATE INDEX IF NOT EXISTS idx_reply_events_date ON reply_events (date);

CREATE TABLE IF NOT EXISTS replaceable_events (
  pubkey TEXT,
  kind INTEGER,
  id TEXT,

  UNIQUE (pubkey, kind)
);
