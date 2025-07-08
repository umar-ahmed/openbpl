-- Table for brands we're protecting
CREATE TABLE brands (
    id SERIAL PRIMARY KEY,             -- Unique number for each brand
    name VARCHAR(255) NOT NULL,        -- Brand name (like "Nike")
    logo_path VARCHAR(500),            -- Where we store their logo
    created_at TIMESTAMP DEFAULT NOW() -- When we added this brand
);

-- Table for suspicious stuff we find
CREATE TABLE threats (
    id SERIAL PRIMARY KEY,            -- Unique number for each threat
    brand_id INTEGER,                 -- Which brand is being copied
    image_url TEXT,                   -- Where we found the suspicious image
    similarity_score FLOAT,           -- How similar it looks (0-100%)
    status VARCHAR(50) DEFAULT 'new', -- new, reviewing, resolved
    created_at TIMESTAMP DEFAULT NOW()
);