CREATE TABLE IF NOT EXISTS rented_cars (
    id bigserial PRIMARY KEY,
    user_id bigint REFERENCES users (id) ON DELETE CASCADE,
    car_id  bigint REFERENCES car (id) ON DELETE CASCADE,
    price integer NOT NULL,
    taking_date date NOT NULL,
    return_date date
);