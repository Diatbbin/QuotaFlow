CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  email VARCHAR NOT NULL UNIQUE,
  username VARCHAR NOT NULL UNIQUE,
  password_hash VARCHAR NOT NULL, 
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ai_tools (
  id BIGSERIAL PRIMARY KEY,
  username VARCHAR NOT NULL,
  tool VARCHAR NOT NULL,
  token_limit BIGINT NOT NULL CHECK (token_limit >= 0),
  tokens_used BIGINT NOT NULL DEFAULT 0 CHECK (tokens_used >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
  CONSTRAINT tokens_used_within_limit CHECK (tokens_used <= token_limit)
);

CREATE TABLE token_transfers (
  id BIGSERIAL PRIMARY KEY,
  from_ai_tool_id BIGINT NOT NULL,
  to_ai_tool_id BIGINT NOT NULL,
  tokens BIGINT NOT NULL CHECK (tokens > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
  CONSTRAINT token_transfers_distinct_ai_tools CHECK (from_ai_tool_id <> to_ai_tool_id)
);

CREATE UNIQUE INDEX ai_tools_user_tool_idx ON ai_tools (username, tool);

CREATE INDEX token_transfers_to_ai_tool_id_idx ON token_transfers (to_ai_tool_id);

CREATE INDEX token_transfers_from_to_idx ON token_transfers (from_ai_tool_id, to_ai_tool_id);

COMMENT ON COLUMN token_transfers.tokens IS 'Unused tokens moved to a colleague';

ALTER TABLE ai_tools ADD FOREIGN KEY (username) REFERENCES users (username);

ALTER TABLE token_transfers ADD FOREIGN KEY (from_ai_tool_id) REFERENCES ai_tools (id) ON DELETE CASCADE;

ALTER TABLE token_transfers ADD FOREIGN KEY (to_ai_tool_id) REFERENCES ai_tools (id) ON DELETE CASCADE;
