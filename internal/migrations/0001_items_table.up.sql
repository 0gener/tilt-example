CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(255)
);