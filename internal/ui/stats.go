package ui

import (
	"fmt"
	"image/color"

	"2d-dungeon/internal/preferences"
	"2d-dungeon/internal/rendering/primitives"
	"2d-dungeon/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type StatsPanel struct {
	scroll float32
}

func (panel *StatsPanel) Draw(player simulation.Player, name string, width, height int32) {
	p := playerDetails(player, name)
	owned := []simulation.PerkCard{}
	for family, rank := range p.Perks {
		if rank > 0 {
			owned = append(owned, simulation.PerkCard{Family: simulation.PerkFamily(family), Rank: rank})
		}
	}
	panelWidth := min(int32(560), width-48)
	blessingsHeight := int32(26)
	if len(owned) > 0 {
		blessingsHeight = int32(len(owned)) * 48
	}
	if p.SpectralShots {
		blessingsHeight += 48
	}
	panelHeight := min(int32(422)+blessingsHeight, height-48)
	x, y := (width-panelWidth)/2, (height-panelHeight)/2
	muted := color.RGBA{160, 170, 180, 255}
	divider := color.RGBA{48, 60, 72, 255}
	rl.DrawRectangle(0, 0, width, height, color.RGBA{6, 10, 17, 190})
	rl.DrawRectangle(x, y, panelWidth, panelHeight, color.RGBA{18, 24, 32, 255})
	rl.DrawRectangleLinesEx(rl.NewRectangle(float32(x), float32(y), float32(panelWidth), float32(panelHeight)), 1, divider)
	rl.DrawRectangle(x, y, panelWidth, 3, primitives.RuneColor)
	left, right := x+28, x+panelWidth-28
	primitives.DrawText("PLAYER STATS", left, y+25, 26, rl.RayWhite)
	primitives.DrawText(preferences.NormalizePlayerName(p.Name), left, y+64, 18, muted)
	level := fmt.Sprintf("LV %d", p.Level)
	primitives.DrawText(level, right-primitives.MeasureText(level, 18), y+64, 18, primitives.RuneColor)
	rl.DrawLine(left, y+98, right, y+98, divider)
	row := func(label, value string, rowY int32) {
		primitives.DrawText(label, left, rowY, 20, muted)
		primitives.DrawText(value, right-primitives.MeasureText(value, 20), rowY, 20, rl.RayWhite)
	}
	primitives.DrawText("COMBAT", left, y+119, 14, primitives.RuneColor)
	row("Damage", fmt.Sprintf("%.2f", p.Damage), y+148)
	row("Attack rate", fmt.Sprintf("%.2f /s", p.AttackRate), y+182)
	primitives.DrawText("MOVEMENT", left, y+228, 14, primitives.RuneColor)
	row("Speed", fmt.Sprintf("%.2f", p.Speed), y+257)
	if player.HasWings {
		row("Shift ability", "STONEWING FLIGHT", y+291)
	} else {
		row("Shift multiplier", fmt.Sprintf("x%.2f", p.Boost), y+291)
	}
	rl.DrawLine(left, y+330, right, y+330, divider)
	primitives.DrawText("BLESSINGS", left, y+349, 14, primitives.RuneColor)
	viewportHeight := max(int32(1), panelHeight-426)
	contentHeight := max(int32(18), blessingsHeight-8)
	maxScroll := float32(max(int32(0), contentHeight-viewportHeight))
	if rl.CheckCollisionPointRec(rl.GetMousePosition(), rl.NewRectangle(float32(x), float32(y), float32(panelWidth), float32(panelHeight))) {
		panel.scroll -= rl.GetMouseWheelMove() * 36
	}
	panel.scroll = max(float32(0), min(maxScroll, panel.scroll))
	rl.BeginScissorMode(left, y+378, right-left, viewportHeight)
	startY := y + 378 - int32(panel.scroll)
	if len(owned) == 0 && !p.SpectralShots {
		primitives.DrawText("No blessings yet", left, startY, 18, muted)
	}
	for i, card := range owned {
		rowY := startY + int32(i)*48
		tint := perkTint(card.Rank)
		label := fmt.Sprintf("%s %s", simulation.PerkName(card.Family), simulation.RankName(card.Rank))
		primitives.DrawText(label, left, rowY, 18, tint)
		primitives.DrawText(card.DescriptionFor(player.HasWings), left, rowY+24, 16, tint)
	}
	if p.SpectralShots {
		rowY := startY + int32(len(owned))*48
		primitives.DrawText("Spectral Shots", left, rowY, 18, primitives.RuneColor)
		primitives.DrawText("Shots release an echo. MP cost +25%.", left, rowY+24, 16, primitives.RuneColor)
	}
	rl.EndScissorMode()
	footer := "[P] / [Esc] Close | World paused"
	if maxScroll > 0 {
		track := rl.NewRectangle(float32(right+12), float32(y+378), 3, float32(viewportHeight))
		rl.DrawRectangleRec(track, divider)
		thumbHeight := max(float32(20), track.Height*track.Height/float32(contentHeight))
		thumbY := track.Y + (track.Height-thumbHeight)*panel.scroll/maxScroll
		rl.DrawRectangleRec(rl.NewRectangle(track.X, thumbY, track.Width, thumbHeight), primitives.RuneColor)
		footer = "Scroll blessings | [P] / [Esc] Close | Paused"
	}
	primitives.DrawText(footer, x+(panelWidth-primitives.MeasureText(footer, 16))/2, y+panelHeight-27, 16, muted)
}
