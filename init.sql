-- JujuDB Database Initialization Script
-- Production password: your-secure-postgres-password
-- Application password: your-secure-app-password

-- Core reference tables
CREATE TABLE IF NOT EXISTS locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS sub_locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    location_id INTEGER NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    UNIQUE(name, location_id)
);

-- Items table (normalized)
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location_id INTEGER REFERENCES locations(id) ON DELETE SET NULL,
    sub_location_id INTEGER REFERENCES sub_locations(id) ON DELETE SET NULL,
    category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    quantity INTEGER DEFAULT 1,
    expiry_date DATE,
    added_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notes TEXT
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_items_name ON items(name);
CREATE INDEX IF NOT EXISTS idx_items_location_id ON items(location_id);
CREATE INDEX IF NOT EXISTS idx_items_sub_location_id ON items(sub_location_id);
CREATE INDEX IF NOT EXISTS idx_items_category_id ON items(category_id);
CREATE INDEX IF NOT EXISTS idx_items_expiry_date ON items(expiry_date);

-- Migrations for existing databases
ALTER TABLE items ADD COLUMN IF NOT EXISTS notes TEXT;

-- Seed reference data
INSERT INTO locations (name) VALUES
('Congélateur'),
('Réfrigérateur'),
('Garde-manger')
ON CONFLICT DO NOTHING;

INSERT INTO categories (name) VALUES
('Viande'),
('Légumes'),
('Desserts'),
('Poisson'),
('Autres')
ON CONFLICT DO NOTHING;

-- Optional sample items using references
INSERT INTO items (name, description, location_id, category_id, quantity, expiry_date, notes)
VALUES
('Steaks de bœuf', 'Steaks de bœuf congelés, 4 pièces',
 (SELECT id FROM locations WHERE name='Congélateur'),
 (SELECT id FROM categories WHERE name='Viande'),
 4, '2024-12-31', NULL),
('Haricots verts', 'Haricots verts surgelés',
 (SELECT id FROM locations WHERE name='Congélateur'),
 (SELECT id FROM categories WHERE name='Légumes'),
 1, '2025-06-30', NULL),
('Glace vanille', 'Bac de glace à la vanille',
 (SELECT id FROM locations WHERE name='Congélateur'),
 (SELECT id FROM categories WHERE name='Desserts'),
 1, '2025-03-15', NULL),
('Saumon fumé', 'Tranches de saumon fumé',
 (SELECT id FROM locations WHERE name='Réfrigérateur'),
 (SELECT id FROM categories WHERE name='Poisson'),
 1, '2024-08-15', NULL),
('Pâtes', 'Paquet de pâtes italiennes',
 (SELECT id FROM locations WHERE name='Garde-manger'),
 (SELECT id FROM categories WHERE name='Autres'),
 2, NULL, NULL),
('Conserves de tomates', 'Boîtes de tomates pelées',
 (SELECT id FROM locations WHERE name='Garde-manger'),
 (SELECT id FROM categories WHERE name='Légumes'),
 6, '2026-01-01', NULL);
