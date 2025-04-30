-- Users table
CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     email VARCHAR(255) NOT NULL UNIQUE,
                                     username VARCHAR(100) NOT NULL UNIQUE,
                                     name VARCHAR(100),
                                     surname VARCHAR(100),
                                     password_hash VARCHAR(255) NOT NULL,
                                     telegram_chat_id BIGINT
);

-- Workspaces table
CREATE TABLE IF NOT EXISTS workspaces (
                                          id SERIAL PRIMARY KEY,
                                          name VARCHAR(255) NOT NULL,
                                          type VARCHAR(50),
                                          hourly_rate NUMERIC NOT NULL,
                                          description TEXT,
                                          capacity INTEGER
);

-- Bookings table
CREATE TABLE IF NOT EXISTS bookings (
                                        id SERIAL PRIMARY KEY,
                                        user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
                                        workspace_id INTEGER REFERENCES workspaces(id) ON DELETE CASCADE,
                                        start_time TIMESTAMP NOT NULL,
                                        end_time TIMESTAMP NOT NULL,
                                        price NUMERIC NOT NULL,
                                        status VARCHAR(20) NOT NULL,
                                        created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                        updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
                                        google_calendar_event_id TEXT
);