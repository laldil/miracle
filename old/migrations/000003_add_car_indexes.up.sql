CREATE INDEX IF NOT EXISTS car_brand_idx ON car USING GIN (to_tsvector('simple', brand));
CREATE INDEX IF NOT EXISTS car_color_idx ON car USING GIN (to_tsvector('simple', color));