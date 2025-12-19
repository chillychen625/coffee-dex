-- CoffeeDex Complete Database Setup
-- Creates database, user, and all necessary tables

-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS coffee_log;

-- Use the database
USE coffee_log;

-- Create coffee_user if it doesn't exist (with error handling)
CREATE USER IF NOT EXISTS 'coffee_user'@'localhost' IDENTIFIED BY 'coffee_pass123';
GRANT ALL PRIVILEGES ON coffee_log.* TO 'coffee_user'@'localhost';
FLUSH PRIVILEGES;

-- Create coffees table (base table that Pokemon tables depend on)
CREATE TABLE IF NOT EXISTS coffees (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    origin VARCHAR(255),
    roaster VARCHAR(255),
    roast_level VARCHAR(50),
    processing_method VARCHAR(100),
    tasting_notes JSON,
    tasting_traits JSON,
    rating INT,
    recipe JSON,
    dripper VARCHAR(100),
    end_time_minutes INT,
    end_time_seconds INT,
    created_at DATETIME,
    updated_at DATETIME
);

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
    FOREIGN KEY (coffee_id) REFERENCES coffees(id) ON DELETE CASCADE,
    FOREIGN KEY (pokemon_id) REFERENCES pokemons(id) ON DELETE CASCADE
);

-- Create unique index to ensure each Pokemon is used only once
-- Note: Index will only be created if it doesn't already exist (handled by CREATE TABLE IF NOT EXISTS)
-- If running setup multiple times, this may fail harmlessly if index exists
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_pokemon ON coffee_pokemon(pokemon_id);

-- Create brewers table (for brewing equipment)
CREATE TABLE IF NOT EXISTS brewers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    pokeball_type VARCHAR(50) NOT NULL,
    created_at DATETIME
);

-- Create brewer_recipes table (for linking brewers to coffee recipes)
CREATE TABLE IF NOT EXISTS brewer_recipes (
    id VARCHAR(36) PRIMARY KEY,
    brewer_id VARCHAR(36) NOT NULL,
    coffee_id VARCHAR(36) NOT NULL,
    created_at DATETIME,
    FOREIGN KEY (brewer_id) REFERENCES brewers(id) ON DELETE CASCADE,
    FOREIGN KEY (coffee_id) REFERENCES coffees(id) ON DELETE CASCADE,
    UNIQUE KEY unique_brewer_coffee (brewer_id, coffee_id)
);

-- Show confirmation
SELECT 'Database setup complete! All tables created successfully.' as status;