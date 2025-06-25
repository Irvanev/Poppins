CREATE TABLE IF NOT EXISTS users (
                       id SERIAL PRIMARY KEY,
                       username TEXT,
                       first_name TEXT,
                       phone TEXT UNIQUE NOT NULL,
                       created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS advertisements (
                                id SERIAL PRIMARY KEY,
                                user_id INT NOT NULL REFERENCES users(id),
                                title TEXT NOT NULL,
                                description TEXT,
                                price BIGINT NOT NULL,
                                photos_urls TEXT[] NOT NULL DEFAULT '{}',
                                address TEXT,
                                archived BOOLEAN NOT NULL DEFAULT FALSE,
                                created_at TIMESTAMP NOT NULL DEFAULT now(),
                                updated_at TIMESTAMP NOT NULL DEFAULT now()
);