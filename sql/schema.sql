-- Coffee Log Database Schema
-- This file contains the complete database schema for the coffee-dex application

-- Coffees table: Stores coffee bean information only
CREATE TABLE IF NOT EXISTS coffees (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    origin VARCHAR(255),
    roaster VARCHAR(255),
    variety VARCHAR(255),
    roast_level VARCHAR(50),
    processing_method VARCHAR(100),
    created_at DATETIME,
    updated_at DATETIME
);

-- Brews table: Stores individual brew/tasting sessions for a coffee
CREATE TABLE IF NOT EXISTS brews (
    id VARCHAR(36) PRIMARY KEY,
    coffee_id VARCHAR(36) NOT NULL,
    tasting_notes JSON,
    tasting_traits JSON,
    rating INT,
    recipe JSON,
    dripper VARCHAR(100),
    end_time_minutes INT,
    end_time_seconds INT,
    created_at DATETIME,
    FOREIGN KEY (coffee_id) REFERENCES coffees(id) ON DELETE CASCADE
);

-- Index for efficient brew queries by coffee
CREATE INDEX idx_brews_coffee_id ON brews(coffee_id);

-- Brewers table: Stores coffee brewing equipment with pokeball sprites
-- Each brewer can have up to 4 standalone recipes stored as JSON
CREATE TABLE IF NOT EXISTS brewers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    pokeball_type VARCHAR(50) NOT NULL,
    recipes JSON,  -- Array of Recipe objects: {id, name, steps[]}
    created_at DATETIME
);

-- Pokemon table: Stores Pokemon data for coffee-to-Pokemon mappings
CREATE TABLE IF NOT EXISTS pokemon (
    id INT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type1 VARCHAR(50) NOT NULL,
    type2 VARCHAR(50),
    sprite_url VARCHAR(255),
    created_at DATETIME
);

-- Coffee-Pokemon mappings: Links coffees to their Pokemon representations
-- Pokemon is generated after 5+ brews of a coffee using aggregated data
CREATE TABLE IF NOT EXISTS coffee_pokemon (
    id VARCHAR(36) PRIMARY KEY,
    coffee_id VARCHAR(36) NOT NULL UNIQUE,
    pokemon_id INT NOT NULL UNIQUE,
    nickname VARCHAR(100),
    level INT DEFAULT 1,
    mapping_confidence REAL,
    llm_description TEXT,
    trait_mapping JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (coffee_id) REFERENCES coffees(id) ON DELETE CASCADE,
    FOREIGN KEY (pokemon_id) REFERENCES pokemon(id)
);
