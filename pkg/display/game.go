package display

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// game implements the ebiten Game interface.
type game struct {
	screen *ebiten.Image
}

var _ ebiten.Game = (*game)(nil) // interface guard

func (g *game) Update() error {
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.screen, nil)
}

func (g *game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// virtual screen size
	return Width, Height
}
