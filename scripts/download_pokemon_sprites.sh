#!/bin/bash

# CoffeeDex Pokemon Sprites Downloader
# Downloads all Gen 1 Pokemon sprites from PokeAPI

SPRITES_DIR="static/pokemon-sprites"
BASE_URL="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon"

# Create sprites directory
mkdir -p "$SPRITES_DIR"

echo "Downloading Gen 1 Pokemon sprites..."

# Download sprites for Pokemon 1-151 (Gen 1)
for i in {1..151}; do
    printf -v padded_id "%03d" $i
    sprite_url="${BASE_URL}/${i}.png"
    output_file="${SPRITES_DIR}/${padded_id}.png"
    
    echo "Downloading Pokemon #$i -> ${output_file}"
    
    if curl -s -f -o "$output_file" "$sprite_url"; then
        echo "  ✓ Downloaded successfully"
    else
        echo "  ✗ Failed to download"
    fi
done

echo "Sprite download complete!"
echo "Total files downloaded: $(ls -1 "$SPRITES_DIR"/*.png 2>/dev/null | wc -l)"