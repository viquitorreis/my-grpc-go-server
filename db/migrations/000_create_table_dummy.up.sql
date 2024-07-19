CREATE TABLE if not exists dummy (
    user_id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)