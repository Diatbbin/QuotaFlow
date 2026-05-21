CREATE TABLE workspaces (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR NOT NULL,
  token_limit BIGINT NOT NULL CHECK (token_limit >= 0),
  tokens_used BIGINT NOT NULL DEFAULT 0 CHECK (tokens_used >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
  CONSTRAINT tokens_used_within_limit CHECK (tokens_used <= token_limit)
);

CREATE TABLE usage_events (
  id BIGSERIAL PRIMARY KEY,
  workspace_id BIGINT NOT NULL,
  tokens BIGINT NOT NULL CHECK (tokens > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT (now())
);

CREATE TABLE token_transfers (
  id BIGSERIAL PRIMARY KEY,
  from_workspace_id BIGINT NOT NULL,
  to_workspace_id BIGINT NOT NULL,
  tokens BIGINT NOT NULL CHECK (tokens > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
  CONSTRAINT token_transfers_distinct_workspaces CHECK (from_workspace_id <> to_workspace_id)
);

CREATE INDEX workspaces_name_idx ON workspaces (name);

CREATE INDEX usage_events_workspace_id_idx ON usage_events (workspace_id);

CREATE INDEX token_transfers_to_workspace_id_idx ON token_transfers (to_workspace_id);

CREATE INDEX token_transfers_from_to_idx ON token_transfers (from_workspace_id, to_workspace_id);

COMMENT ON COLUMN usage_events.tokens IS 'can only be negative (spent)';
COMMENT ON COLUMN token_transfers.tokens IS 'Unused tokens moved to a colleague';

ALTER TABLE usage_events ADD FOREIGN KEY (workspace_id) REFERENCES workspaces (id);

ALTER TABLE token_transfers ADD FOREIGN KEY (from_workspace_id) REFERENCES workspaces (id);

ALTER TABLE token_transfers ADD FOREIGN KEY (to_workspace_id) REFERENCES workspaces (id);
