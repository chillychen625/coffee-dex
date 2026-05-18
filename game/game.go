package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	InternalWidth  = 480
	InternalHeight = 320
	WindowScale    = 2
	TileSize       = 16
)

// Game implements ebiten.Game.
type Game struct {
	svc    *Services
	world  *World
	player *Player
	scene  SceneID

	// Interaction prompt state
	promptBuilding *Building
}

// Run initialises and starts the game loop.
func Run(svc *Services) error {
	ebiten.SetWindowSize(InternalWidth*WindowScale, InternalHeight*WindowScale)
	ebiten.SetWindowTitle("CoffeeDex")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	world := NewWorld()
	g := &Game{
		svc:    svc,
		world:  world,
		player: NewPlayer(14, 16, world), // start on the main path
		scene:  SceneOverworld,
	}
	return ebiten.RunGame(g)
}

// Update is called every tick (~60fps).
func (g *Game) Update() error {
	if g.scene == SceneOverworld {
		entered := g.player.Update()
		if entered != SceneOverworld {
			g.scene = entered
		}
		// Update interaction prompt
		g.promptBuilding = g.world.BuildingAt(g.player.TileX(), g.player.TileY())
	} else {
		// Placeholder: pressing Escape returns to the overworld
		if isEscJustPressed() {
			g.scene = SceneOverworld
		}
	}
	return nil
}

// Draw renders the current scene.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.scene == SceneOverworld {
		g.drawOverworld(screen)
	} else {
		g.drawPlaceholderScene(screen, g.scene)
	}
}

// Layout returns the internal canvas dimensions.
func (g *Game) Layout(_, _ int) (int, int) {
	return InternalWidth, InternalHeight
}

func (g *Game) drawOverworld(screen *ebiten.Image) {
	// Draw tiles
	for row := 0; row < MapRows; row++ {
		for col := 0; col < MapCols; col++ {
			tile := g.world.Tiles[row][col]
			c, ok := TileColors[tile]
			if !ok {
				c = color.RGBA{R: 80, G: 80, B: 80, A: 255}
			}
			vector.DrawFilledRect(screen,
				float32(col*TileSize), float32(row*TileSize),
				TileSize, TileSize, c, false)
		}
	}

	// Draw building roofs / labels
	for i := range g.world.Buildings {
		b := &g.world.Buildings[i]
		vector.DrawFilledRect(screen,
			float32(b.TileX*TileSize), float32(b.TileY*TileSize),
			float32(b.W*TileSize), float32(b.H*TileSize),
			b.RoofColor, false)
		ebitenutil.DebugPrintAt(screen, b.Name,
			b.TileX*TileSize+2, b.TileY*TileSize+2)
	}

	// Draw player
	g.player.Draw(screen)

	// Draw interaction prompt
	if g.promptBuilding != nil {
		drawDialogBox(screen, fmt.Sprintf("Enter %s? [Z]", g.promptBuilding.Name))
	}
}

func (g *Game) drawPlaceholderScene(screen *ebiten.Image, id SceneID) {
	screen.Fill(color.RGBA{R: 20, G: 20, B: 40, A: 255})
	names := map[SceneID]string{
		SceneRoastery:   "Roastery",
		SceneWarehouse:  "Bean Warehouse",
		ScenePokemonLab: "Pokemon Lab",
		SceneTrophyRoom: "Trophy Room",
	}
	name := names[id]
	ebitenutil.DebugPrintAt(screen, name+" — coming soon", 16, 16)
	ebitenutil.DebugPrintAt(screen, "[Esc] to leave", 16, 32)
}

// drawDialogBox draws a Gen-3-style bottom dialog box.
func drawDialogBox(screen *ebiten.Image, text string) {
	boxH := float32(40)
	boxY := float32(InternalHeight) - boxH - 4
	// Border
	vector.DrawFilledRect(screen, 4, boxY, float32(InternalWidth)-8, boxH,
		color.RGBA{R: 240, G: 240, B: 240, A: 255}, false)
	vector.DrawFilledRect(screen, 8, boxY+4, float32(InternalWidth)-16, boxH-8,
		color.RGBA{R: 30, G: 30, B: 60, A: 255}, false)
	ebitenutil.DebugPrintAt(screen, text, 14, int(boxY)+10)
}

func isEscJustPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyEscape)
}
