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
  id            SERIAL PRIMARY KEY,
  title         TEXT NOT NULL,
  status        TEXT NOT NULL,  -- 'uploaded' | 'processing' | 'processed' | 'failed'
  uploaded_at   TIMESTAMP NOT NULL,
  processed_at  TIMESTAMP NULL,
  origin_url    TEXT NOT NULL,
  processed_url TEXT,
  votes         INT NOT NULL DEFAULT 0,
  user_id       INT NOT NULL
);


CREATE INDEX IF NOT EXISTS idx_anb_video     ON videos(id);
