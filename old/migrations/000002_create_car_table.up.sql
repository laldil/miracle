CREATE TABLE IF NOT EXISTS car (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    brand text NOT NULL,
    description text NOT NULL,
    color text NOT NULL,
    year integer NOT NULL,
    price integer NOT NULL,
    is_used boolean NOT NULL DEFAULT false,
    owner_id bigint REFERENCES users(id) ON DELETE CASCADE NOT NULL
);