-- Coffee Log Database Schema
-- This file contains the complete database schema for the coffee-dex application

-- Coffees table: Stores coffee entries with brewing details
CREATE TABLE IF NOT EXISTS coffees (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    origin VARCHAR(255),
    roaster VARCHAR(255),
    variety VARCHAR(255),
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
CREATE TABLE IF NOT EXISTS coffee_pokemon (
    coffee_id VARCHAR(36) PRIMARY KEY,
    pokemon_id INT NOT NULL,
    nickname VARCHAR(100),
    created_at DATETIME,
    FOREIGN KEY (coffee_id) REFERENCES coffees(id) ON DELETE CASCADE,
    FOREIGN KEY (pokemon_id) REFERENCES pokemon(id)
);

-- DEPRECATED TABLES (kept for backward compatibility, will be removed in future)
-- These tables are no longer used in the application

-- Legacy brewer_recipes table (DEPRECATED)
-- Previously used for coffee-based recipes, now replaced by standalone recipes in brewers.recipes JSON column
CREATE TABLE IF NOT EXISTS brewer_recipes (
    id VARCHAR(36) PRIMARY KEY,
    brewer_id VARCHAR(36) NOT NULL,
    coffee_id VARCHAR(36) NOT NULL,
    created_at DATETIME,
    FOREIGN KEY (brewer_id) REFERENCES brewers(id) ON DELETE CASCADE,
    FOREIGN KEY (coffee_id) REFERENCES coffees(id) ON DELETE CASCADE,
    UNIQUE KEY unique_brewer_coffee (brewer_id, coffee_id)
);