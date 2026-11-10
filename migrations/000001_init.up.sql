-- Todos los ids son UUID v4 generados por la aplicación (Go, github.com/google/uuid);
-- la DB no genera ids (no DEFAULT gen_random_uuid()).
-- Todos los timestamps los genera la aplicación en Go (time.Now().UTC());
-- la DB no define DEFAULT now() (SPECS-001.4.7).
CREATE TABLE channels (
    id              UUID PRIMARY KEY,
    number          INTEGER NOT NULL UNIQUE CHECK (number > 0),
    name            TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 100),
    logo            TEXT,
    schedule_url    TEXT,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    segment_seconds INTEGER NOT NULL DEFAULT 5 CHECK (segment_seconds BETWEEN 2 AND 10),
    sources         JSONB NOT NULL DEFAULT '[null,null,null,null,null]'::jsonb
                    CHECK (jsonb_array_length(sources) = 5),  -- array fijo de 5 slots; cada slot: objeto Source o null (SPECS-001.5.1)
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL
);

CREATE TABLE config (
    key   TEXT PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
-- Semillas con los defaults de SPECS-001.5.3:
INSERT INTO config (key, value) VALUES
    ('neighbors', '2'),
    ('startup_channels', '3'),
    ('current_channel_ttl_minutes', '30'),
    ('buffer_minutes', '10'),
    ('uptime_threshold', '0.95');

CREATE TABLE uptime_samples (
    id          UUID PRIMARY KEY,
    channel_id  UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    source_id   UUID,                           -- id de la source embebida en channels.sources; NULL = muestra global del canal. Sin FK: es referencia lógica al JSON
    source_slot INTEGER CHECK (source_slot BETWEEN 1 AND 5),
    ts          TIMESTAMPTZ NOT NULL,   -- generado por la app (rodritv-playlist, SPECS-002)
    ok_count    INTEGER NOT NULL,   -- muestras buenas de la ventana de 30 s
    total_count INTEGER NOT NULL    -- muestras totales (6 con sample cada 5 s)
);
CREATE INDEX idx_uptime_channel_ts ON uptime_samples(channel_id, ts DESC);

CREATE TABLE process_logs (
    id         UUID PRIMARY KEY,
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    source_id  UUID,                            -- referencia lógica a la source embebida (sin FK)
    ts         TIMESTAMPTZ NOT NULL,            -- generado por la app (rodritv-playlist, SPECS-002)
    process    TEXT NOT NULL,                  -- 'ffmpeg' | 'yt-dlp' | 'chromium' | 'mixer' | 'extractor'
    level      TEXT NOT NULL CHECK (level IN ('info', 'warn', 'error')),
    message    TEXT NOT NULL                   -- incluye comando, exit code y stderr relevante
);
CREATE INDEX idx_process_logs_channel_ts ON process_logs(channel_id, ts DESC);
