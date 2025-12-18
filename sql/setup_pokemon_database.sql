-- CoffeeDex Pokemon Database Setup
-- Creates the Pokemon tables needed for the coffee-Pokemon integration

-- Create pokemons reference table
CREATE TABLE IF NOT EXISTS pokemons (
    id INT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    sprite_path VARCHAR(255) NOT NULL,
    base_stats JSON NOT NULL,
    description TEXT
);

-- Create coffee_pokemon mapping table
CREATE TABLE IF NOT EXISTS coffee_pokemon (
    id VARCHAR(36) PRIMARY KEY,
    coffee_id VARCHAR(36) NOT NULL,
    pokemon_id INT NOT NULL,
    nickname VARCHAR(100),
    level INT DEFAULT 1,
    mapping_confidence REAL,
    llm_description TEXT,
    trait_mapping JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (coffee_id) REFERENCES coffees(id),
    FOREIGN KEY (pokemon_id) REFERENCES pokemons(id)
);

-- Create unique index to ensure each Pokemon is used only once
CREATE UNIQUE INDEX idx_unique_pokemon ON coffee_pokemon(pokemon_id);

-- Show confirmation
SELECT 'Pokemon tables created successfully!' as status;