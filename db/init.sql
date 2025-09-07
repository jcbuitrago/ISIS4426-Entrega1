CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    surname         VARCHAR(50) UNIQUE NOT NULL,
    lastname        VARCHAR(50) UNIQUE NOT NULL,
    city            VARCHAR(50) UNIQUE NOT NULL,
    email           VARCHAR(50) UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    video_url     TEXT
);

CREATE INDEX IF NOT EXISTS idx_anb_user     ON users(id);

CREATE TABLE IF NOT EXISTS videos (
  id         SERIAL PRIMARY KEY,
  title      TEXT NOT NULL,
  url        TEXT NOT NULL,
  status     TEXT NOT NULL CHECK (status IN ('uploaded','processing','processed','failed')),
  user_id    INTEGER,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_anb_video     ON videos(id);
