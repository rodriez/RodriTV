CREATE TABLE channels (
    id VARCHAR PRIMARY KEY,
    name VARCHAR NOT NULL,
    list_position INT NOT NULL,
    status VARCHAR PRIMARY KEY,
    signals JSON NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_channel_list ON channels(list_position);