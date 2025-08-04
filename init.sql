-- JujuDB Database Initialization Script
-- Production password: your-secure-postgres-password
-- Application password: your-secure-app-password

-- Create the items table
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(100) NOT NULL DEFAULT 'Congélateur',
    category VARCHAR(100),
    quantity INTEGER DEFAULT 1,
    expiry_date DATE,
    added_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_items_name ON items(name);
CREATE INDEX IF NOT EXISTS idx_items_location ON items(location);
CREATE INDEX IF NOT EXISTS idx_items_category ON items(category);
CREATE INDEX IF NOT EXISTS idx_items_expiry_date ON items(expiry_date);

-- Insert some sample data for testing
INSERT INTO items (name, description, location, category, quantity, expiry_date) VALUES
('Steaks de bœuf', 'Steaks de bœuf congelés, 4 pièces', 'Congélateur', 'Viande', 4, '2024-12-31'),
('Haricots verts', 'Haricots verts surgelés', 'Congélateur', 'Légumes', 1, '2025-06-30'),
('Glace vanille', 'Bac de glace à la vanille', 'Congélateur', 'Desserts', 1, '2025-03-15'),
('Saumon fumé', 'Tranches de saumon fumé', 'Réfrigérateur', 'Poisson', 1, '2024-08-15'),
('Pâtes', 'Paquet de pâtes italiennes', 'Garde-manger', 'Autres', 2, NULL),
('Conserves de tomates', 'Boîtes de tomates pelées', 'Garde-manger', 'Légumes', 6, '2026-01-01');
