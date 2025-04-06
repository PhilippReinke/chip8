package display

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	Width, Height = 64, 32
)

// Display abstracts Chip8 display.
//
// It can render a scaled screen with 64x32 pixel and handles keyboard input.
type Display struct {
	scale  float64
	screen *ebiten.Image
}

// New creates new display instance of given scale.
func New[T float64 | int](scale T) *Display {
	return &Display{
		scale:  max(float64(scale), 0),
		screen: ebiten.NewImage(Width, Height),
	}
}

// Run starts application.
func (d *Display) Run() error {
	ebiten.SetWindowSize(d.scaledSize())

	return ebiten.RunGame(&game{
		screen: d.screen,
	})
}

// UpdateScreen updates screen content.
func (d *Display) UpdateScreen(screen [Height][Width]byte) {
	for r := range screen {
		for c, b := range screen[r] {
			col := color.Black
			if b != 0 {
				col = color.White
			}
			d.screen.Set(c, r, col)
		}
	}
}

func (d *Display) scaledSize() (int, int) {
	if d.scale == 0 {
		return int(Width * 10), int(Height * 10)
	}

	return int(Width * d.scale), int(Height * d.scale)
}
