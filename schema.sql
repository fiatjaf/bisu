CREATE TABLE IF NOT EXISTS pubkey_relays (
  pubkey text NOT NULL,
  relay text NOT NULL,
  last_fetched_attempt int NOT NULL DEFAULT 0,
  last_fetched_success int NOT NULL DEFAULT 0,
  last_hint_nprofile int NOT NULL DEFAULT 0,
  last_hint_nip05 int NOT NULL DEFAULT 0,
  last_hint_tag int NOT NULL DEFAULT 0,
  last_hint_kind3 int NOT NULL DEFAULT 0,
  last_nip65_outbox int NOT NULL DEFAULT 0,
  last_nip65_inbox int NOT NULL DEFAULT 0,
  last_kind3_outbox int NOT NULL DEFAULT 0,
  last_kind3_inbox int NOT NULL DEFAULT 0,

  UNIQUE (pubkey, relay)
);
