-- +goose Up

CREATE TABLE IF NOT EXISTS history(
    id SERIAL PRIMARY KEY,
    transcription_type VARCHAR(4) NOT NULL,
    type_price VARCHAR(10) NOT NULL,
    price DECIMAL(18, 8) NOT NULL,
    volume DECIMAL(18, 8) NOT NULL,
    amount DECIMAL(18, 8) NOT NULL,
    time_stamp_order BIGINT NOT NULL
);
