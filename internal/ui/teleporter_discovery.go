package ui

import (
	"dungeon-raylib-go/internal/rendering"
	"dungeon-raylib-go/internal/rendering/primitives"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const teleporterDiscoveryDuration = float32(9)

type TeleporterDiscovery struct {
	seen      bool
	remaining float32
}

func (p *TeleporterDiscovery) Update(dt float32, discovered, powered bool) {
	if powered {
		p.seen, p.remaining = true, 0
		return
	}
	p.remaining = max(0, p.remaining-dt)
	if !p.seen && discovered {
		p.seen, p.remaining = true, teleporterDiscoveryDuration
		rendering.PlayHintSound()
	}
}

func (p *TeleporterDiscovery) Draw(width, height int32) {
	if p.remaining <= 0 {
		return
	}
	alpha := min(float32(1), (teleporterDiscoveryDuration-p.remaining)/.35, p.remaining)
	panelWidth := min(int32(580), width-48)
	lines := []string{
		"What is this... a teleportation circle?",
		"It won't activate. Maybe something is missing...",
	}
	fontSize := int32(20)
	for fontSize > 10 && primitives.MeasureText(lines[1], fontSize) > panelWidth-40 {
		fontSize--
	}
	panelHeight := int32(64) + 2*fontSize
	x, y := (width-panelWidth)/2, int32(108)
	if x < 340 {
		y = 270
	}
	y = min(y, max(int32(16), height-panelHeight-86))
	rl.DrawRectangle(x, y, panelWidth, panelHeight, rl.Fade(rl.NewColor(15, 22, 29, 255), .94*alpha))
	rl.DrawRectangle(x, y, 3, panelHeight, rl.Fade(primitives.RuneColor, alpha))
	primitives.DrawText("A DORMANT CIRCLE", x+20, y+16, 16, rl.Fade(primitives.RuneColor, alpha))
	for i, line := range lines {
		primitives.DrawText(line, x+20, y+42+int32(i)*(fontSize+6), fontSize, rl.Fade(rl.RayWhite, alpha))
	}
}
