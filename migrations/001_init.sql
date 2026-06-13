-- DARX: initial schema
-- 001_init.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

--  users
CREATE TABLE users (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address      VARCHAR(42) NOT NULL UNIQUE,
    username            VARCHAR(32) UNIQUE,
    avatar_url          TEXT,
    credits             INT NOT NULL DEFAULT 0 CHECK (credits >= 0),
    total_wins          INT NOT NULL DEFAULT 0,
    total_losses        INT NOT NULL DEFAULT 0,
    rating              INT NOT NULL DEFAULT 1000,
    active_character_id UUID,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_wallet  ON users (wallet_address);
CREATE INDEX idx_users_rating  ON users (rating DESC);

-- characters
CREATE TABLE characters (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(64) NOT NULL,
    rarity      VARCHAR(16) NOT NULL CHECK (rarity IN ('common','rare','epic','legendary')),
    asset_url   TEXT NOT NULL,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

--  user_characters
CREATE TABLE user_characters (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id),
    acquired_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, character_id)
);

CREATE INDEX idx_user_characters_user ON user_characters (user_id);

ALTER TABLE users
    ADD CONSTRAINT fk_active_character
    FOREIGN KEY (active_character_id) REFERENCES characters(id);

--duels
CREATE TABLE duels (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player1_id    UUID NOT NULL REFERENCES users(id),
    player2_id    UUID REFERENCES users(id),
    asset         VARCHAR(10) NOT NULL CHECK (asset IN ('BTC','ETH','SOL')),
    timeframe     SMALLINT NOT NULL CHECK (timeframe IN (1,3,5)),
    credits_bet   INT NOT NULL CHECK (credits_bet > 0),
    status        VARCHAR(16) NOT NULL DEFAULT 'waiting'
                    CHECK (status IN ('waiting','active','finished','cancelled')),
    direction_p1  VARCHAR(8) CHECK (direction_p1 IN ('UP','DOWN','FLAT')),
    direction_p2  VARCHAR(8) CHECK (direction_p2 IN ('UP','DOWN','FLAT')),
    price_open    NUMERIC(18,8),
    price_close   NUMERIC(18,8),
    winner_id     UUID REFERENCES users(id),
    raise_count   SMALLINT NOT NULL DEFAULT 0,
    starts_at     TIMESTAMPTZ,
    ends_at       TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_duels_player1  ON duels (player1_id);
CREATE INDEX idx_duels_player2  ON duels (player2_id);
CREATE INDEX idx_duels_status   ON duels (status) WHERE status IN ('waiting','active');
CREATE INDEX idx_duels_created  ON duels (created_at DESC);

-- items
CREATE TABLE items (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             VARCHAR(64) NOT NULL,
    type             VARCHAR(32) NOT NULL CHECK (type IN ('skin','emote','effect','banner')),
    rarity           VARCHAR(16) NOT NULL CHECK (rarity IN ('common','rare','epic','legendary')),
    credit_cost      INT NOT NULL CHECK (credit_cost >= 0),
    asset_url        TEXT NOT NULL,
    is_nft           BOOLEAN NOT NULL DEFAULT FALSE,
    contract_address VARCHAR(42),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- user_items
CREATE TABLE user_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id     UUID NOT NULL REFERENCES items(id),
    token_id    VARCHAR(78),
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, item_id)
);

CREATE INDEX idx_user_items_user ON user_items (user_id);

--  achievements
CREATE TABLE achievements (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug             VARCHAR(64) NOT NULL UNIQUE,
    title            VARCHAR(128) NOT NULL,
    description      TEXT,
    icon_url         TEXT,
    condition_type   VARCHAR(32) NOT NULL,
    condition_value  INT NOT NULL DEFAULT 1,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- user_achievements
CREATE TABLE user_achievements (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id),
    unlocked_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, achievement_id)
);

CREATE INDEX idx_user_achievements_user ON user_achievements (user_id);

--  seasons
CREATE TABLE seasons (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    number     INT NOT NULL UNIQUE,
    name       VARCHAR(64) NOT NULL,
    starts_at  TIMESTAMPTZ NOT NULL,
    ends_at    TIMESTAMPTZ NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (ends_at > starts_at)
);

CREATE UNIQUE INDEX idx_seasons_active ON seasons (is_active) WHERE is_active = TRUE;

--season_standings
CREATE TABLE season_standings (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    season_id     UUID NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating_points INT NOT NULL DEFAULT 0,
    rank          INT,
    wins          INT NOT NULL DEFAULT 0,
    losses        INT NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (season_id, user_id)
);

CREATE INDEX idx_standings_season_rating ON season_standings (season_id, rating_points DESC);

--updated_at trigger
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_standings_updated_at
    BEFORE UPDATE ON season_standings
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
