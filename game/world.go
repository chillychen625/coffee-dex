package game

import "image/color"

// Tile types
const (
	TileGrass    = 0
	TilePath     = 1
	TileBuilding = 2
	TileWater    = 3
	TileFence    = 4
)

// TileColors used for placeholder rendering until real sprites are in place.
var TileColors = map[int]color.RGBA{
	TileGrass:    {R: 104, G: 176, B: 72, A: 255},
	TilePath:     {R: 200, G: 168, B: 112, A: 255},
	TileBuilding: {R: 180, G: 120, B: 80, A: 255},
	TileWater:    {R: 80, G: 160, B: 224, A: 255},
	TileFence:    {R: 160, G: 120, B: 64, A: 255},
}

// Building represents an interactable structure on the map.
type Building struct {
	// TileX/TileY: top-left corner of the building on the tile grid
	TileX, TileY int
	// W/H: dimensions in tiles
	W, H int
	// Name shown in the interaction prompt
	Name string
	// Scene to open when the player interacts
	SceneID SceneID
	// RoofColor for placeholder rendering
	RoofColor color.RGBA
	// DoorTileX/Y: the tile the player must be adjacent to (facing the door)
	DoorTileX, DoorTileY int
}

// SceneID identifies which scene/screen to open.
type SceneID int

const (
	SceneOverworld  SceneID = iota
	SceneRoastery           // log a brew + recent brews + brewers
	SceneWarehouse          // add/manage coffees
	ScenePokemonLab         // pokédex
	SceneTrophyRoom         // gym badges + stats
)

// MapCols and MapRows define the overworld tile dimensions.
// 480px / 16px = 30 cols, 320px / 16px = 20 rows
const (
	MapCols = 30
	MapRows = 20
)

// World holds the tile map and buildings.
type World struct {
	Tiles     [MapRows][MapCols]int
	Buildings []Building
}

// NewWorld builds the initial town layout.
func NewWorld() *World {
	w := &World{}
	w.buildBuildings() // defines buildings first so buildTileMap can stamp them
	return w
}

func (w *World) buildTileMap() {
	// Fill with grass
	for row := 0; row < MapRows; row++ {
		for col := 0; col < MapCols; col++ {
			w.Tiles[row][col] = TileGrass
		}
	}

	// Horizontal main path (row 14-15, full width)
	for col := 0; col < MapCols; col++ {
		w.Tiles[14][col] = TilePath
		w.Tiles[15][col] = TilePath
	}

	// Vertical path connecting buildings (col 14-15)
	for row := 0; row < MapRows; row++ {
		w.Tiles[row][14] = TilePath
		w.Tiles[row][15] = TilePath
	}

	// Mark building footprints on the tile map
	for _, b := range w.Buildings {
		for row := b.TileY; row < b.TileY+b.H; row++ {
			for col := b.TileX; col < b.TileX+b.W; col++ {
				if row >= 0 && row < MapRows && col >= 0 && col < MapCols {
					w.Tiles[row][col] = TileBuilding
				}
			}
		}
	}
}

func (w *World) buildBuildings() {
	w.Buildings = []Building{
		{
			Name:      "Roastery",
			TileX:     2, TileY: 2, W: 6, H: 5,
			SceneID:   SceneRoastery,
			RoofColor: color.RGBA{R: 200, G: 80, B: 60, A: 255},
			DoorTileX: 4, DoorTileY: 7,
		},
		{
			Name:      "Bean Warehouse",
			TileX:     18, TileY: 2, W: 6, H: 5,
			SceneID:   SceneWarehouse,
			RoofColor: color.RGBA{R: 60, G: 120, B: 200, A: 255},
			DoorTileX: 20, DoorTileY: 7,
		},
		{
			Name:      "Pokemon Lab",
			TileX:     2, TileY: 10, W: 6, H: 4,
			SceneID:   ScenePokemonLab,
			RoofColor: color.RGBA{R: 80, G: 180, B: 120, A: 255},
			DoorTileX: 4, DoorTileY: 14,
		},
		{
			Name:      "Trophy Room",
			TileX:     18, TileY: 10, W: 6, H: 4,
			SceneID:   SceneTrophyRoom,
			RoofColor: color.RGBA{R: 200, G: 160, B: 40, A: 255},
			DoorTileX: 20, DoorTileY: 14,
		},
	}

	// Stamp building footprints into the tile map
	w.buildTileMap()
}

// IsWalkable returns true if the player can step on the given tile.
func (w *World) IsWalkable(tileX, tileY int) bool {
	if tileX < 0 || tileX >= MapCols || tileY < 0 || tileY >= MapRows {
		return false
	}
	return w.Tiles[tileY][tileX] != TileBuilding && w.Tiles[tileY][tileX] != TileWater && w.Tiles[tileY][tileX] != TileFence
}

// BuildingAt returns the building whose door tile the player is standing on, if any.
func (w *World) BuildingAt(tileX, tileY int) *Building {
	for i := range w.Buildings {
		b := &w.Buildings[i]
		if b.DoorTileX == tileX && b.DoorTileY == tileY {
			return b
		}
	}
	return nil
}
