CREATE TABLE users (
    id BYTEA PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL
);

CREATE TABLE credentials (
    id BYTEA NOT NULL,
    user_id BYTEA NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE CASCADE,
    counter INT NOT NULL,
    cose_key BYTEA NOT NULL,
    registered_at TIMESTAMP WITH TIME ZONE,
    loggedin_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY(id, user_id)
);
