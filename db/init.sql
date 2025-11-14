-- USERS
CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    first_name    VARCHAR(50)  NOT NULL,
    last_name     VARCHAR(50)  NOT NULL,
    city          VARCHAR(50)  NOT NULL,
    country       VARCHAR(50)  NOT NULL,
    avatar_url    TEXT         NULL,
    email         VARCHAR(120) NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_email ON users(email);

-- VIDEOS
CREATE TABLE IF NOT EXISTS videos (
  id            SERIAL PRIMARY KEY,
  title         TEXT        NOT NULL,
  status        TEXT        NOT NULL,        -- 'uploaded' | 'processing' | 'processed' | 'failed'
  uploaded_at   TIMESTAMP   NOT NULL DEFAULT NOW(),
  processed_at  TIMESTAMP   NULL,
  origin_url    VARCHAR(512)        NOT NULL,
  processed_url VARCHAR(512)        NULL,
  thumb_url     VARCHAR(512)        NULL,
  votes         INT         NOT NULL DEFAULT 0,
  user_id       INT         NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_videos_user_id   ON videos(user_id);
CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status);

-- VOTES (1 por usuario por video)
CREATE TABLE IF NOT EXISTS votes (
  video_id  INT NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
  user_id   INT NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  PRIMARY KEY (video_id, user_id)
);

-- JOB STATUS
CREATE TABLE IF NOT EXISTS job_status (
  job_id        VARCHAR(50) PRIMARY KEY,
  status        TEXT NOT NULL,
  created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
  expires_at    TIMESTAMP    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_job_status_expires ON job_status(expires_at);