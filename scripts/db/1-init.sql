CREATE TABLE orders_config (
    pack_sizes integer[] NOT NULL
);

INSERT INTO orders_config (pack_sizes) VALUES ('{250, 500, 1000, 2000, 5000}');
