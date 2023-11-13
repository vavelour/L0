CREATE TABLE orders
(
    id  varchar PRIMARY KEY NOT NULL,
    data jsonb               NOT NULL
);

ALTER TABLE orders OWNER TO valera;