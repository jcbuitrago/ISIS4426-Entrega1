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