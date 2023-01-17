CREATE TABLE IF NOT EXISTS relay_scores (
  pubkey TEXT NOT NULL,
  url TEXT NOT NULL,
  score INT NOT NULL,

  UNIQUE (pubkey, url)
);

CREATE TABLE IF NOT EXISTS events (
  id TEXT NOT NULL,
  body BLOB NOT NULL,

  PRIMARY KEY (id) ON CONFLICT IGNORE
);

CREATE TABLE IF NOT EXISTS profile_events (
  pubkey TEXT NOT NULL,
  id TEXT NOT NULL,
  date INTEGER NOT NULL,

  PRIMARY KEY (pubkey, id) ON CONFLICT IGNORE
);

CREATE INDEX IF NOT EXISTS idx_profile_events_date ON profile_events (date);

CREATE TABLE IF NOT EXISTS reply_events (
  root TEXT NOT NULL,
  id TEXT NOT NULL,
  date INTEGER NOT NULL,

  PRIMARY KEY (root, id) ON CONFLICT IGNORE
);

CREATE INDEX IF NOT EXISTS idx_reply_events_date ON reply_events (date);

CREATE TABLE IF NOT EXISTS replaceable_events (
  pubkey TEXT NOT NULL,
  kind INTEGER NOT NULL,
  id TEXT NOT NULL,

  PRIMARY KEY (pubkey, kind) ON CONFLICT IGNORE
);

CREATE TABLE IF NOT EXISTS follows (
  pubkey TEXT PRIMARY KEY
);
