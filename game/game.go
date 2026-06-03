package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	InternalWidth  = 480
	InternalHeight = 320
	WindowScale    = 2
)

type SceneID int

const (
	SceneMenu        SceneID = iota
	SceneRoastery
	SceneWarehouse
	ScenePokemonLab
	SceneTrophyRoom
)

// Scene is implemented by each screen in the app.
type Scene interface {
	Update() SceneID
	Draw(*ebiten.Image)
	OnEnter(*Services)
}

type Game struct {
	svc     *Services
	sceneID SceneID
	scenes  map[SceneID]Scene
}

func Run(svc *Services) error {
	ebiten.SetWindowSize(InternalWidth*WindowScale, InternalHeight*WindowScale)
	ebiten.SetWindowTitle("CoffeeDex")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	menu := NewMenuScene()
	roastery := NewRoasteryScene()
	warehouse := NewWarehouseScene()
	lab := NewPokemonLabScene()
	trophy := NewTrophyRoomScene()

	g := &Game{
		svc:     svc,
		sceneID: SceneMenu,
		scenes: map[SceneID]Scene{
			SceneMenu:       menu,
			SceneRoastery:   roastery,
			SceneWarehouse:  warehouse,
			ScenePokemonLab: lab,
			SceneTrophyRoom: trophy,
		},
	}
	menu.OnEnter(svc)
	return ebiten.RunGame(g)
}

func (g *Game) Update() error {
	next := g.scenes[g.sceneID].Update()
	if next != g.sceneID {
		g.sceneID = next
		g.scenes[g.sceneID].OnEnter(g.svc)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.scenes[g.sceneID].Draw(screen)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return InternalWidth, InternalHeight
}

func isKeyJustPressed(k ebiten.Key) bool {
	return inpututil.IsKeyJustPressed(k)
}
