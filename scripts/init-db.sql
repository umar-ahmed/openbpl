-- Create a basic users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert a test user
INSERT INTO users (email, name) 
VALUES ('test@openbpl.local', 'Test User')
ON CONFLICT (email) DO NOTHING;

-- Create a basic threats table  
CREATE TABLE IF NOT EXISTS threats (
    id SERIAL PRIMARY KEY,
    url VARCHAR(500) NOT NULL,
    status VARCHAR(50) DEFAULT 'new',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Insert a test threat
INSERT INTO threats (url, status) 
VALUES ('https://fake-brand-site.com', 'investigating')
ON CONFLICT DO NOTHING;